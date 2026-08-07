// Package butane defines Go structs that represent the Flatcar Butane YAML
// configuration format. See: https://coreos.github.io/butane/config-flatcar-v1_0/
package butane

// Config is the top-level Butane document.
type Config struct {
	Variant  string   `yaml:"variant"`
	Version  string   `yaml:"version"`
	Passwd   *Passwd  `yaml:"passwd,omitempty"`
	Storage  *Storage `yaml:"storage,omitempty"`
	Systemd  *Systemd `yaml:"systemd,omitempty"`
}

// Passwd holds user and group definitions in Butane.
type Passwd struct {
	Users  []User  `yaml:"users,omitempty"`
	Groups []Group `yaml:"groups,omitempty"`
}

// User represents a user entry in Butane.
type User struct {
	Name              string   `yaml:"name"`
	Groups            []string `yaml:"groups,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
	HomeDir           string   `yaml:"home_dir,omitempty"`
	NoCreateHome      bool     `yaml:"no_create_home,omitempty"`
	System            bool     `yaml:"system,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	ShouldExist       *bool    `yaml:"should_exist,omitempty"`
	Gecos             string   `yaml:"gecos,omitempty"`
	PasswordHash      string   `yaml:"password_hash,omitempty"`
}

// Group represents a system group entry in Butane.
type Group struct {
	Name        string `yaml:"name"`
	Gid         *int   `yaml:"gid,omitempty"`
	System      bool   `yaml:"system,omitempty"`
	ShouldExist *bool  `yaml:"should_exist,omitempty"`
}

// Storage holds file system definitions in Butane.
type Storage struct {
	Files []File `yaml:"files,omitempty"`
}

// File represents a file entry in Butane storage.
type File struct {
	Path     string      `yaml:"path"`
	Mode     *int        `yaml:"mode,omitempty"`
	Overwrite bool       `yaml:"overwrite,omitempty"`
	Contents *FileContent `yaml:"contents,omitempty"`
	User     *FileUser   `yaml:"user,omitempty"`
	Group    *FileGroup  `yaml:"group,omitempty"`
	Append   []FileAppend `yaml:"append,omitempty"`
}

// FileContent holds the content specification for a Butane file.
type FileContent struct {
	Inline string `yaml:"inline,omitempty"`
}

// FileUser holds owner user information for a Butane file.
type FileUser struct {
	Name string `yaml:"name,omitempty"`
}

// FileGroup holds owner group information for a Butane file.
type FileGroup struct {
	Name string `yaml:"name,omitempty"`
}

// FileAppend holds content to append to a Butane file.
type FileAppend struct {
	Inline string `yaml:"inline,omitempty"`
}

// Systemd holds systemd unit definitions in Butane.
type Systemd struct {
	Units []Unit `yaml:"units,omitempty"`
}

// Unit represents a systemd unit entry in Butane.
type Unit struct {
	Name     string `yaml:"name"`
	Contents string `yaml:"contents,omitempty"`
	Enabled  *bool  `yaml:"enabled,omitempty"`
	Mask     bool   `yaml:"mask,omitempty"`
}
