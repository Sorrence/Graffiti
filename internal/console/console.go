// Package console provides consistently formatted terminal status messages.
package console

import "fmt"

const (
	reset = "\033[0m"
	blue  = "\033[34m"
	green = "\033[32m"
	red   = "\033[31m"
)

// Status formats a terminal message with a colored status marker.
// Supported markers are: [!] (blue), [+] (green), and [-] (red).
func Status(marker, format string, args ...any) string {
	color := ""
	switch marker {
	case "[!]":
		color = blue
	case "[+]":
		color = green
	case "[-]":
		color = red
	}

	return fmt.Sprintf("%s%s%s %s", color, marker, reset, fmt.Sprintf(format, args...))
}
