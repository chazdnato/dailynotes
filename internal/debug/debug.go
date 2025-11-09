package debug

import "fmt"

// Enabled controls whether debug output is printed
var Enabled = false

// Printf prints debug output if debug mode is enabled
func Printf(format string, args ...interface{}) {
	if Enabled {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}
