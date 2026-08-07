// Package cloudconfig defines Go structs that represent a subset of the
// cloud-init cloud-config YAML format. Only fields relevant to ClusterAPI
// worker node provisioning are included.
package cloudconfig

// Config is the top-level cloud-config document.
type Config struct {
	Users          []User        `yaml:"users,omitempty"`
	Groups         []Group       `yaml:"groups,omitempty"`
	WriteFiles     []File        `yaml:"write_files,omitempty"`
	RunCMD         []interface{} `yaml:"runcmd,omitempty"`
	SystemdUnits   []SystemdUnit `yaml:"systemd,omitempty"`
	CACerts        CACerts       `yaml:"ca_certs,omitempty"`
	Hostname       string        `yaml:"hostname,omitempty"`
	ManageEtcHosts string        `yaml:"manage_etc_hosts,omitempty"`
}

// User represents a user entry in cloud-config.
type User struct {
	Name              string   `yaml:"name"`
	Groups            string   `yaml:"groups,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
	HomeDir           string   `yaml:"homedir,omitempty"`
	NoCreateHome      bool     `yaml:"no_create_home,omitempty"`
	System            bool     `yaml:"system,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh-authorized-keys,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty"`
	Gecos             string   `yaml:"gecos,omitempty"`
	LockPasswd        bool     `yaml:"lock_passwd,omitempty"`
	PasswordHash      string   `yaml:"passwd,omitempty"`
}

// Group represents a system group entry in cloud-config.
type Group struct {
	Name    string   `yaml:"name"`
	Members []string `yaml:"members,omitempty"`
	Gid     int      `yaml:"gid,omitempty"`
	System  bool     `yaml:"system,omitempty"`
}

// File represents a file to be written by cloud-config.
type File struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Permissions string `yaml:"permissions,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Encoding    string `yaml:"encoding,omitempty"`
	Append      bool   `yaml:"append,omitempty"`
}

// SystemdUnit represents a systemd unit in cloud-config.
type SystemdUnit struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty"`
	Mask    bool   `yaml:"mask,omitempty"`
}

// CACerts represents certificate authority certificates in cloud-config.
type CACerts struct {
	Trusted []string `yaml:"trusted,omitempty"`
}
