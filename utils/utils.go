package utils

import "golang.org/x/sys/windows"

// IsDebuggerPresent Returns true if a debugger is detected on the system.
func IsDebuggerPresent() bool {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	ret, _, _ := isDebuggerPresent.Call()
	return ret != 0
}
