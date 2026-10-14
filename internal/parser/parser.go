package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Result represents the top-level structure of enum4linux-ng -oJ/-oY output.
type Result struct {
	Target   Target             `json:"target" yaml:"target"`
	Services map[string]Service `json:"services" yaml:"services"`
	SMB      map[string]any     `json:"smb_dialects" yaml:"smb_dialects"`
	OSInfo   map[string]any     `json:"os_info" yaml:"os_info"`
	Users    map[string]User    `json:"users" yaml:"users"`
	Groups   map[string]Group   `json:"groups" yaml:"groups"`
	Shares   map[string]Share   `json:"shares" yaml:"shares"`
	Policy   map[string]any     `json:"policy" yaml:"policy"`
	Errors   map[string]any     `json:"errors" yaml:"errors"`

	SessionsPossible    *bool `json:"sessions_possible" yaml:"sessions_possible"`
	NullSessionPossible *bool `json:"null_session_possible" yaml:"null_session_possible"`
}

// Target contains basic information about the target.
type Target struct {
	Host      string `json:"host" yaml:"host"`
	Workgroup string `json:"workgroup" yaml:"workgroup"`
}

// Service contains port and availability information for services such as SMB/RPC.
type Service struct {
	Port       int  `json:"port" yaml:"port"`
	Accessible bool `json:"accessible" yaml:"accessible"`
}

// User represents a domain or local user found during enumeration.
type User struct {
	Username    string `json:"username" yaml:"username"`
	Name        string `json:"name" yaml:"name"`
	ACB         string `json:"acb" yaml:"acb"`
	Description string `json:"description" yaml:"description"`
}

// Group represents a group found during enumeration.
type Group struct {
	Groupname   string `json:"groupname" yaml:"groupname"`
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`
}

// Share represents a discovered SMB share.
type Share struct {
	Comment string `json:"comment" yaml:"comment"`
	Type    string `json:"type" yaml:"type"`
	Mapping string `json:"mapping" yaml:"mapping"`
	Listing string `json:"listing" yaml:"listing"`
	Writing string `json:"writing" yaml:"writing"`
}

// ParseFile detects the supplied file's extension (.json, .yaml, or .yml)
// and decodes it into a Result.
func ParseFile(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	return Parse(data, filepath.Ext(path))
}

// Parse decodes raw data and its extension (".json", ".yaml", or ".yml")
// into a Result. It is the file-independent form of ParseFile, useful for
// tests or data read from stdin or the network.
func Parse(data []byte, ext string) (*Result, error) {
	var r Result

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("json parse: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("yaml parse: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported extension: %s (expected: .json, .yaml, .yml)", ext)
	}

	return &r, nil
}
