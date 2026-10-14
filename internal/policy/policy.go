package policy

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// PasswordPolicy represents a domain password policy extracted from enum4linux-ng output.
type PasswordPolicy struct {
	MinLength         int
	HistoryLength     int
	ComplexityOn      bool
	LockoutThreshold  int // 0 = no account lockout policy
	LockoutDuration   time.Duration
	ObservationWindow time.Duration
}

// Config contains the final values used for filtering and timing after
// command-line overrides are applied.
type Config struct {
	MinLength        int
	LockoutThreshold int
	Delay            time.Duration
}

// Extract builds a PasswordPolicy from parser.Result.Policy.
func Extract(raw map[string]any) PasswordPolicy {
	p := PasswordPolicy{}

	// The top-level key (for example, "Domain Password Information") can vary
	// by version, so inspect nested maps and normalize their keys.
	flat := map[string]any{}
	for _, v := range raw {
		if m, ok := v.(map[string]any); ok {
			for k2, v2 := range m {
				flat[normalizeKey(k2)] = v2
			}
		}
	}

	p.MinLength = asInt(flat["min_pwd_length"])
	p.HistoryLength = asInt(flat["pwd_history_length"])
	p.LockoutThreshold = asInt(flat["lockout_threshold"])
	p.ComplexityOn = hasComplexity(flat["pwd_properties"])
	p.LockoutDuration = asDuration(flat["lockout_duration"])
	p.ObservationWindow = asDuration(flat["lockout_observation_window"])

	return p
}

// BuildConfig creates the final Config from PasswordPolicy and overrides it
// only with flags explicitly supplied by the user (using fs.Changed).
// Precedence: CLI flag > enum4linux-ng output > calculated default.
func BuildConfig(fs *pflag.FlagSet, pol PasswordPolicy) Config {
	cfg := Config{
		MinLength:        pol.MinLength,
		LockoutThreshold: pol.LockoutThreshold,
		Delay:            computeDelay(pol),
	}

	if fs.Changed("min-length") {
		if v, err := fs.GetInt("min-length"); err == nil {
			cfg.MinLength = v
		}
	}
	if fs.Changed("lockout-threshold") {
		if v, err := fs.GetInt("lockout-threshold"); err == nil {
			cfg.LockoutThreshold = v
		}
	}
	if fs.Changed("delay") {
		if v, err := fs.GetDuration("delay"); err == nil {
			cfg.Delay = v
		}
	}

	return cfg
}

// computeDelay calculates a safe delay between attempts from the lockout
// threshold and observation window, leaving a 20% safety margin.
func computeDelay(p PasswordPolicy) time.Duration {
	if p.LockoutThreshold <= 0 || p.ObservationWindow <= 0 {
		return 0 // no account lockout policy; no wait is needed
	}
	const safeMargin = 0.8
	perAttempt := time.Duration(float64(p.ObservationWindow) / float64(p.LockoutThreshold) * safeMargin)
	return perAttempt
}

func normalizeKey(k string) string {
	k = strings.ToLower(k)
	k = strings.ReplaceAll(k, " ", "_")
	return k
}

func asInt(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	}
	return 0
}

var durRe = regexp.MustCompile(`(-?\d+)\s*(day|hour|minute|second)s?`)

// asDuration parses free-form durations such as "30 minutes" and
// "1 day, 5 hours". It returns 0 for "None" or "Not set".
func asDuration(v any) time.Duration {
	s, ok := v.(string)
	if !ok {
		return 0
	}
	if strings.EqualFold(s, "none") || strings.EqualFold(s, "not set") {
		return 0
	}
	var total time.Duration
	for _, m := range durRe.FindAllStringSubmatch(s, -1) {
		n, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "day":
			total += time.Duration(n) * 24 * time.Hour
		case "hour":
			total += time.Duration(n) * time.Hour
		case "minute":
			total += time.Duration(n) * time.Minute
		case "second":
			total += time.Duration(n) * time.Second
		}
	}
	return total
}

func hasComplexity(v any) bool {
	switch t := v.(type) {
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && strings.Contains(strings.ToLower(s), "complex") {
				return true
			}
		}
	case string:
		return strings.Contains(strings.ToLower(t), "complex")
	}
	return false
}
