# Graffiti

Graffiti is a Go tool for running more controlled SMB password attempts during **authorized Active Directory security assessments**. It reads the target and password policy from `enum4linux-ng` JSON/YAML output, filters a password list to policy-compatible entries, skips users that approach the lockout threshold, and records successful credentials.

> Use only on systems you own or have explicit written permission to assess. Unauthorized password attempts can cause account lockouts, service disruption, and legal consequences.

## Features

- Parses JSON, YAML, and YML results produced by `enum4linux-ng`.
- Extracts the minimum length, complexity requirement, and lockout threshold from the password policy.
- Filters out unsuitable passwords before attempting authentication.
- Reduces risk by skipping a user one attempt before the lockout threshold.
- Calculates a delay between attempts from the lockout observation window.
- Appends successful username/password pairs to a selected output file.

## Requirements

- Go 1.23 or later
- A reachable SMB service (TCP/445)
- `enum4linux-ng` JSON or YAML output for the target
- User and password lists

## Installation

```bash
git clone https://github.com/sorrence/graffiti.git
cd graffiti
go build -o graffiti .
```

Optionally download dependencies in advance:

```bash
go mod download
```

## Usage

First, generate enum4linux-ng output for a target within your authorized scope:

```bash
enum4linux-ng -A -oJ enum.json <target>
```

Then pass the user and password lists to Graffiti:

```bash
./graffiti \
  --input enum.json \
  --userlist users.txt \
  --passlist passwords.txt \
  --output successful-logins.txt
```

The tool connects to the target host specified in `enum.json` over TCP/445. Each successful login is appended to the output file in `username:password` format.

### Options

| Option | Description |
| --- | --- |
| `-i`, `--input` | Required. enum4linux-ng JSON/YAML/YML output. |
| `-u`, `--userlist` | Required. File containing one username per line. |
| `-p`, `--passlist` | Required. File containing one password per line. |
| `-o`, `--output` | File for successful logins. Default: `output.txt`. |
| `--min-length` | Overrides the detected minimum password length. |
| `--lockout-threshold` | Overrides the detected account lockout threshold. |
| `--delay` | Overrides the delay between attempts, e.g. `30s` or `2m`. |
| `-v`, `--verbose` | Also prints failed attempts. |

Example: explicitly set the delay required by your assessment plan:

```bash
./graffiti -i enum.yaml -u users.txt -p passwords.txt --delay 45s -v
```

## Password policy and rate control

By default, Graffiti uses the following fields from enum4linux-ng output:

- Minimum password length
- Password complexity
- Account lockout threshold
- Lockout observation window

When complexity is enabled, a password must contain at least three categories: uppercase letters, lowercase letters, digits, and symbols. When the delay is calculated automatically, Graffiti applies a 20% safety margin based on the observation window and lockout threshold. Explicit CLI options take precedence over detected values.

These safeguards are not a security guarantee. Do not run the tool in a production environment before assessing scope, authorization, maintenance windows, and the account-lockout recovery plan.

## Development

```bash
go test ./...
go vet ./...
```

## License

This repository does not currently include a license file. Confirm licensing terms with the project owner before distributing or reusing it.
