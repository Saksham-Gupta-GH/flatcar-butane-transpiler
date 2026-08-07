package transpiler_test

import (
	"testing"

	"github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/pkg/cloudconfig"
	"github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/pkg/transpiler"
)

// TestTranspileUsers verifies that cloud-config users are correctly mapped to
// Butane passwd users, including SSH key and group conversion.
func TestTranspileUsers(t *testing.T) {
	cfg := &cloudconfig.Config{
		Users: []cloudconfig.User{
			{
				Name:              "core",
				Groups:            "sudo, docker",
				Shell:             "/bin/bash",
				SSHAuthorizedKeys: []string{"ssh-rsa AAAA...key1", "ssh-rsa AAAA...key2"},
			},
		},
	}

	out, warnings, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if out.Passwd == nil {
		t.Fatal("expected Passwd block in Butane config, got nil")
	}
	if len(out.Passwd.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(out.Passwd.Users))
	}

	u := out.Passwd.Users[0]
	if u.Name != "core" {
		t.Errorf("expected user name 'core', got %q", u.Name)
	}
	if u.Shell != "/bin/bash" {
		t.Errorf("expected shell '/bin/bash', got %q", u.Shell)
	}
	if len(u.SSHAuthorizedKeys) != 2 {
		t.Errorf("expected 2 SSH keys, got %d", len(u.SSHAuthorizedKeys))
	}
	// cloud-config groups comma string should become a slice
	if len(u.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d: %v", len(u.Groups), u.Groups)
	}
}

// TestTranspileFiles verifies that write_files entries are correctly mapped to
// Butane storage files with correct permissions and owner parsing.
func TestTranspileFiles(t *testing.T) {
	cfg := &cloudconfig.Config{
		WriteFiles: []cloudconfig.File{
			{
				Path:        "/etc/myapp/config.toml",
				Content:     "[server]\nport = 8080\n",
				Permissions: "0600",
				Owner:       "myapp:myapp",
			},
		},
	}

	out, _, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	if out.Storage == nil {
		t.Fatal("expected Storage block in Butane config, got nil")
	}
	if len(out.Storage.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(out.Storage.Files))
	}

	f := out.Storage.Files[0]
	if f.Path != "/etc/myapp/config.toml" {
		t.Errorf("expected path '/etc/myapp/config.toml', got %q", f.Path)
	}
	if f.Mode == nil {
		t.Fatal("expected file mode to be set, got nil")
	}
	// 0600 octal = 384 decimal
	if *f.Mode != 0600 {
		t.Errorf("expected mode 0600 (384), got %d", *f.Mode)
	}
	if f.User == nil || f.User.Name != "myapp" {
		t.Errorf("expected user 'myapp', got %v", f.User)
	}
	if f.Group == nil || f.Group.Name != "myapp" {
		t.Errorf("expected group 'myapp', got %v", f.Group)
	}
	if f.Contents == nil || f.Contents.Inline == "" {
		t.Error("expected file contents to be set")
	}
}

// TestTranspileSystemdUnits verifies that cloud-config systemd units are
// correctly mapped to Butane systemd units.
func TestTranspileSystemdUnits(t *testing.T) {
	cfg := &cloudconfig.Config{
		SystemdUnits: []cloudconfig.SystemdUnit{
			{
				Name:    "myapp.service",
				Content: "[Unit]\nDescription=My App\n[Service]\nExecStart=/usr/bin/myapp\n",
				Enabled: true,
			},
		},
	}

	out, _, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	if out.Systemd == nil {
		t.Fatal("expected Systemd block in Butane config, got nil")
	}
	if len(out.Systemd.Units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(out.Systemd.Units))
	}

	u := out.Systemd.Units[0]
	if u.Name != "myapp.service" {
		t.Errorf("expected unit name 'myapp.service', got %q", u.Name)
	}
	if u.Enabled == nil || !*u.Enabled {
		t.Error("expected unit to be enabled")
	}
}

// TestTranspileGroups verifies that cloud-config groups are correctly mapped
// to Butane passwd groups.
func TestTranspileGroups(t *testing.T) {
	cfg := &cloudconfig.Config{
		Groups: []cloudconfig.Group{
			{Name: "docker", System: true},
		},
	}

	out, _, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	if out.Passwd == nil || len(out.Passwd.Groups) != 1 {
		t.Fatal("expected 1 group in Butane config")
	}
	g := out.Passwd.Groups[0]
	if g.Name != "docker" {
		t.Errorf("expected group name 'docker', got %q", g.Name)
	}
	if !g.System {
		t.Error("expected group to be a system group")
	}
}

// TestTranspileVariantVersion verifies the Butane output always has the
// correct variant and version fields.
func TestTranspileVariantVersion(t *testing.T) {
	out, _, err := transpiler.Transpile(&cloudconfig.Config{})
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	if out.Variant != "flatcar" {
		t.Errorf("expected variant 'flatcar', got %q", out.Variant)
	}
	if out.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", out.Version)
	}
}

// TestTranspileRunCMDWarning verifies that unsupported runcmd fields produce
// a warning and do not cause an error.
func TestTranspileRunCMDWarning(t *testing.T) {
	cfg := &cloudconfig.Config{
		RunCMD: []interface{}{"echo hello"},
	}
	_, warnings, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("expected no error for runcmd, got: %v", err)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for runcmd, got none")
	}
}

// TestTranspileAppendFile verifies that a write_files entry with append: true
// generates a Butane append section rather than overwriting contents.
func TestTranspileAppendFile(t *testing.T) {
	cfg := &cloudconfig.Config{
		WriteFiles: []cloudconfig.File{
			{
				Path:    "/etc/hosts",
				Content: "127.0.0.1 myhost\n",
				Append:  true,
			},
		},
	}
	out, _, err := transpiler.Transpile(cfg)
	if err != nil {
		t.Fatalf("Transpile() returned unexpected error: %v", err)
	}
	f := out.Storage.Files[0]
	if f.Contents != nil {
		t.Error("expected Contents to be nil for append file")
	}
	if len(f.Append) == 0 {
		t.Error("expected Append section to be set for append file")
	}
}
