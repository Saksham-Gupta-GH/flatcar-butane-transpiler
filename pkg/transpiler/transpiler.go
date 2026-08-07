// Package transpiler converts cloud-init cloud-config YAML documents into
// Flatcar Butane YAML documents.
//
// The scope is deliberately narrow: only the subset of cloud-config fields
// needed to support ClusterAPI worker node provisioning is handled.
// Unsupported fields are ignored with a warning rather than causing a fatal
// error, so that partial configs are still useful.
package transpiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/pkg/butane"
	"github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/pkg/cloudconfig"
)

const (
	butaneVariant = "flatcar"
	butaneVersion = "1.0.0"
)

// Transpile converts a parsed cloud-config document into a Butane document.
// Warnings about unsupported fields are returned alongside a valid (possibly
// partial) Butane config.
func Transpile(cfg *cloudconfig.Config) (*butane.Config, []string, error) {
	out := &butane.Config{
		Variant: butaneVariant,
		Version: butaneVersion,
	}

	var warnings []string

	// --- Users ---
	if len(cfg.Users) > 0 {
		if out.Passwd == nil {
			out.Passwd = &butane.Passwd{}
		}
		for _, u := range cfg.Users {
			bu, w := transpileUser(u)
			out.Passwd.Users = append(out.Passwd.Users, bu)
			warnings = append(warnings, w...)
		}
	}

	// --- Groups ---
	if len(cfg.Groups) > 0 {
		if out.Passwd == nil {
			out.Passwd = &butane.Passwd{}
		}
		for _, g := range cfg.Groups {
			bg, w := transpileGroup(g)
			out.Passwd.Groups = append(out.Passwd.Groups, bg)
			warnings = append(warnings, w...)
		}
	}

	// --- Files ---
	if len(cfg.WriteFiles) > 0 {
		if out.Storage == nil {
			out.Storage = &butane.Storage{}
		}
		for _, f := range cfg.WriteFiles {
			bf, w, err := transpileFile(f)
			if err != nil {
				return nil, warnings, fmt.Errorf("transpiling file %q: %w", f.Path, err)
			}
			out.Storage.Files = append(out.Storage.Files, bf)
			warnings = append(warnings, w...)
		}
	}

	// --- Systemd Units ---
	if len(cfg.SystemdUnits) > 0 {
		if out.Systemd == nil {
			out.Systemd = &butane.Systemd{}
		}
		for _, u := range cfg.SystemdUnits {
			bu := transpileSystemdUnit(u)
			out.Systemd.Units = append(out.Systemd.Units, bu)
		}
	}

	// --- CA Certs ---
	if len(cfg.CACerts.Trusted) > 0 {
		if out.Storage == nil {
			out.Storage = &butane.Storage{}
		}
		for i, cert := range cfg.CACerts.Trusted {
			path := fmt.Sprintf("/etc/ssl/certs/cloud-config-ca-%d.pem", i)
			trueVal := true
			out.Storage.Files = append(out.Storage.Files, butane.File{
				Path:      path,
				Overwrite: true,
				Contents:  &butane.FileContent{Inline: cert},
				// Butane mode 0644 = 420 decimal
				Mode: intPtr(0644),
			})
			_ = trueVal
		}
	}

	// --- Unsupported but commonly used fields ---
	if len(cfg.RunCMD) > 0 {
		warnings = append(warnings,
			"runcmd is not supported in Butane; consider converting commands to a systemd oneshot unit")
	}
	if cfg.Hostname != "" {
		warnings = append(warnings,
			"hostname is not supported in Butane; set it via a systemd unit or kernel cmdline")
	}

	return out, warnings, nil
}

// transpileUser maps a cloud-config user to a Butane user.
func transpileUser(u cloudconfig.User) (butane.User, []string) {
	var warnings []string

	bu := butane.User{
		Name:              u.Name,
		Shell:             u.Shell,
		HomeDir:           u.HomeDir,
		NoCreateHome:      u.NoCreateHome,
		System:            u.System,
		Gecos:             u.Gecos,
		SSHAuthorizedKeys: u.SSHAuthorizedKeys,
		PasswordHash:      u.PasswordHash,
	}

	// cloud-config groups is a comma-separated string; Butane expects a slice
	if u.Groups != "" {
		parts := strings.Split(u.Groups, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				bu.Groups = append(bu.Groups, trimmed)
			}
		}
	}

	if u.Sudo != "" {
		warnings = append(warnings,
			fmt.Sprintf("user %q: sudo field is not supported in Butane; configure sudo via a write_files entry for /etc/sudoers.d/", u.Name))
	}

	return bu, warnings
}

// transpileGroup maps a cloud-config group to a Butane group.
func transpileGroup(g cloudconfig.Group) (butane.Group, []string) {
	var warnings []string
	bg := butane.Group{
		Name:   g.Name,
		System: g.System,
	}
	if g.Gid != 0 {
		bg.Gid = intPtr(g.Gid)
	}
	if len(g.Members) > 0 {
		warnings = append(warnings,
			fmt.Sprintf("group %q: members field is not supported in Butane; add users to groups via the user's groups field", g.Name))
	}
	return bg, warnings
}

// transpileFile maps a cloud-config write_files entry to a Butane storage file.
func transpileFile(f cloudconfig.File) (butane.File, []string, error) {
	var warnings []string

	bf := butane.File{
		Path: f.Path,
	}

	// Permissions (e.g. "0644") → integer mode
	if f.Permissions != "" {
		mode, err := strconv.ParseInt(strings.TrimPrefix(f.Permissions, "0"), 8, 32)
		if err != nil {
			return butane.File{}, warnings,
				fmt.Errorf("invalid permissions %q: %w", f.Permissions, err)
		}
		bf.Mode = intPtr(int(mode))
	}

	// Owner: "user:group" or just "user"
	if f.Owner != "" {
		parts := strings.SplitN(f.Owner, ":", 2)
		if parts[0] != "" {
			bf.User = &butane.FileUser{Name: parts[0]}
		}
		if len(parts) == 2 && parts[1] != "" {
			bf.Group = &butane.FileGroup{Name: parts[1]}
		}
	}

	// Encoding: cloud-config supports base64; Butane uses data URLs
	if f.Encoding == "b64" || f.Encoding == "base64" {
		warnings = append(warnings,
			fmt.Sprintf("file %q: base64 encoding detected; converting to Butane data URL format", f.Path))
		bf.Contents = &butane.FileContent{Inline: "data:text/plain;charset=utf-8;base64," + f.Content}
	} else {
		if f.Append {
			bf.Append = []butane.FileAppend{{Inline: f.Content}}
		} else {
			bf.Contents = &butane.FileContent{Inline: f.Content}
		}
	}

	return bf, warnings, nil
}

// transpileSystemdUnit maps a cloud-config systemd unit to a Butane unit.
func transpileSystemdUnit(u cloudconfig.SystemdUnit) butane.Unit {
	bu := butane.Unit{
		Name:     u.Name,
		Contents: u.Content,
		Mask:     u.Mask,
	}
	if u.Enabled {
		bu.Enabled = boolPtr(true)
	}
	return bu
}

func intPtr(i int) *int   { return &i }
func boolPtr(b bool) *bool { return &b }
