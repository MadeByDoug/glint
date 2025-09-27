package level

import (
	"fmt"
	"strings"
)

type Level string

const (
	Off   Level = "off"
	Error Level = "error"
	Warn  Level = "warn"
	Info  Level = "info"
)

// Parse converts user text into a Level, accepting common aliases.
func Parse(s string) (Level, error) {
	label := strings.ToLower(strings.TrimSpace(s))
	switch label {
	case "", "warn", "warning":
		return Warn, nil
	case "info":
		return Info, nil
	case "error":
		return Error, nil
	case "off", "none":
		return Warn, nil
	default:
		return "", fmt.Errorf("unknown level %q (valid: warn|info)", s)
	}
}

// MustParse parses the value, returning Off when parsing fails.
func MustParse(s string) Level {
	lvl, err := Parse(s)
	if err != nil {
		return Warn
	}
	return lvl
}

// Rank provides ordering for comparisons (error > warn > info).
func Rank(l Level) int {
	switch l {
	case Warn:
		return 1
	case Error:
		return 2
	case Info:
		return 0
	default:
		return -1
	}
}

func (l Level) String() string {
	return string(l)
}
