# flatcar-butane-transpiler

[![CI](https://github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/actions/workflows/ci.yml/badge.svg)](https://github.com/Saksham-Gupta-GH/flatcar-butane-transpiler/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Saksham-Gupta-GH/flatcar-butane-transpiler)](https://goreportcard.com/report/github.com/Saksham-Gupta-GH/flatcar-butane-transpiler)
[![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen)](https://github.com/Saksham-Gupta-GH/flatcar-butane-transpiler)

A command-line tool written in Go that converts **cloud-init cloud-config YAML** files into **Flatcar Butane YAML** configuration files.

This project was built as a proof-of-concept prototype for the [CNCF - Flatcar Container Linux: Cloud-Init to Butane YAML config transpiler (2026 Term 3)](https://github.com/flatcar/Flatcar/issues/2226) LFX Mentorship proposal.

## Background

[Flatcar Container Linux](https://www.flatcar.org/) uses [Butane](https://coreos.github.io/butane/) and Ignition for provisioning. However, much of the wider Kubernetes ecosystem (including ClusterAPI) still produces `cloud-config` (cloud-init) YAML, which makes it harder to adopt Flatcar in environments where cloud-config is already the established norm.

This tool bridges that gap by automatically translating cloud-config documents into their Butane equivalents, enabling ClusterAPI worker node provisioning workflows to work seamlessly with Flatcar.

## Supported Features

The scope is deliberately narrow, covering the minimum set of features needed for ClusterAPI worker node provisioning:

| cloud-config field | Butane equivalent              | Notes                                        |
|--------------------|-------------------------------|----------------------------------------------|
| `users`            | `passwd.users`                | Including `ssh-authorized-keys` → `ssh_authorized_keys`, comma-separated `groups` → slice |
| `groups`           | `passwd.groups`               | Including `system` and `gid` fields          |
| `write_files`      | `storage.files`               | Including permissions, owner parsing, append mode, base64 encoding |
| `systemd`          | `systemd.units`               | Including `enabled` and `mask` fields        |
| `ca_certs`         | `storage.files`               | Written to `/etc/ssl/certs/`                 |
| `runcmd`           | `systemd.units` & `storage.files` | Auto-transpiled into a `oneshot` systemd unit executing a generated bash script |

### Explicitly Not Supported

The following cloud-config fields are not representable in Butane. The transpiler emits a **warning** for each one instead of failing, so partial configs remain useful:

- `hostname` — set via kernel cmdline or a systemd unit
- `manage_etc_hosts`
- `sudo` on user entries — write a file to `/etc/sudoers.d/` instead

## Installation

```bash
git clone https://github.com/Saksham-Gupta-GH/flatcar-butane-transpiler.git
cd flatcar-butane-transpiler
go build -o transpiler .
```

Requires Go 1.22 or later.

## Usage

```bash
# Output to stdout
./transpiler -input cloud-config.yaml

# Output to a file
./transpiler -input cloud-config.yaml -output butane.yaml

# Strict mode: exit non-zero if any unsupported fields are encountered
./transpiler -input cloud-config.yaml -strict
```

## Example

**Input** (`examples/worker-node.cloud-config.yaml`):

```yaml
#cloud-config
users:
  - name: core
    groups: sudo, docker
    shell: /bin/bash
    ssh-authorized-keys:
      - ssh-rsa AAAAB3NzaC1yc2E... user@example.com

write_files:
  - path: /etc/myapp/config.toml
    permissions: "0644"
    owner: root:root
    content: |
      [server]
      port = 8080

systemd:
  - name: myapp.service
    enabled: true
    content: |
      [Unit]
      Description=My Application
      [Service]
      ExecStart=/usr/bin/myapp
      [Install]
      WantedBy=multi-user.target
```

**Output** (Butane YAML):

```yaml
variant: flatcar
version: 1.0.0
passwd:
    users:
        - name: core
          groups:
            - sudo
            - docker
          shell: /bin/bash
          ssh_authorized_keys:
            - ssh-rsa AAAAB3NzaC1yc2E... user@example.com
storage:
    files:
        - path: /etc/myapp/config.toml
          mode: 420
          contents:
            inline: |
              [server]
              port = 8080
          user:
            name: root
          group:
            name: root
systemd:
    units:
        - name: myapp.service
          contents: |
            [Unit]
            Description=My Application
            [Service]
            ExecStart=/usr/bin/myapp
            [Install]
            WantedBy=multi-user.target
          enabled: true
```

## Running Tests

```bash
# Run all unit tests
make test

# Run the transpiler on the example file
make run-example
```

## Project Structure

```
.
├── main.go                          # CLI entry point
├── pkg/
│   ├── cloudconfig/
│   │   ├── types.go                 # cloud-config struct definitions
│   │   └── parse.go                 # YAML parser (handles #cloud-config header)
│   ├── butane/
│   │   └── types.go                 # Butane struct definitions
│   └── transpiler/
│       ├── transpiler.go            # Core field mapping logic
│       └── transpiler_test.go       # Unit tests
├── examples/
│   └── worker-node.cloud-config.yaml
├── .github/workflows/ci.yml         # CI: build + test on Go 1.22 & 1.23
└── Makefile
```

## Known Limitations

- `sudo` grants are not translated; write a file to `/etc/sudoers.d/` instead.
- Group `members` lists are not supported in Butane; assign groups via the `user.groups` field.
- Only the `flatcar` Butane variant (`v1.0.0`) is produced.

## References

- [Butane Configuration Specification (Flatcar v1.0)](https://coreos.github.io/butane/config-flatcar-v1_0/)
- [cloud-init cloud-config Reference](https://cloudinit.readthedocs.io/en/latest/reference/examples.html)
- [Flatcar Container Linux](https://www.flatcar.org/)
- [ClusterAPI](https://cluster-api.sigs.k8s.io/)

