package wordlist

import (
	"bufio"
	"fmt"
	"os"
	"unicode"

	"github.com/sorrence/graffiti/internal/console"
	"github.com/sorrence/graffiti/internal/policy"
)

// Filter filters passwords from in according to the policy in cfg and returns
// the valid entries as a []string. Reading line by line keeps memory use low,
// even for large wordlists (only accepted entries are retained in memory).
func Filter(in string, cfg policy.Config, complexityOn bool) ([]string, error) {
	f, err := os.Open(in)
	if err != nil {
		return nil, fmt.Errorf("could not open wordlist: %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Increase the buffer to support wordlists with lines longer than 64 KB.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var result []string
	total := 0
	for sc.Scan() {
		total++
		pw := sc.Text()
		if pw == "" {
			continue
		}
		if cfg.MinLength > 0 && len(pw) < cfg.MinLength {
			continue
		}
		if complexityOn && !meetsComplexity(pw) {
			continue
		}
		result = append(result, pw)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	fmt.Println(console.Status("[+]", "Read %d passwords; %d match the policy.", total, len(result)))
	return result, nil
}

// meetsComplexity applies Active Directory's standard complexity rule:
// at least three of the uppercase, lowercase, digit, and symbol categories.
func meetsComplexity(pw string) bool {
	var upper, lower, digit, symbol bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			symbol = true
		}
	}
	count := 0
	for _, b := range []bool{upper, lower, digit, symbol} {
		if b {
			count++
		}
	}
	return count >= 3
}
