// internal/ux/ux.go
package ux

import (
	"fmt"
	"os"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

var verboseMode = false

// SetVerbose enables/disables debug output
func SetVerbose(v bool) { verboseMode = v }
func IsVerbose() bool   { return verboseMode }

// Info prints cyan informational message
func Info(msg string) {
	fmt.Printf("%sℹ️  %s%s\n", Cyan, msg, Reset)
}

// Success prints green success message
func Success(msg string) {
	fmt.Printf("%s✅ %s%s\n", Green, msg, Reset)
}

// Error prints red error to stderr
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "%s❌ %s%s\n", Red, msg, Reset)
}

// Warning prints yellow warning
func Warning(msg string) {
	fmt.Printf("%s⚠️  %s%s\n", Yellow, msg, Reset)
}

// Verbose prints only if --verbose/-v flag is set
func Verbose(msg string) {
	if verboseMode {
		fmt.Printf("%s🔍 %s%s\n", Cyan, msg, Reset)
	}
}

// Tip prints beginner-friendly hint
func Tip(msg string) {
	fmt.Printf("%s💡 %s%s\n", Yellow, msg, Reset)
}
