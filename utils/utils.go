package utils

import "golang.org/x/sys/windows"

// IsDebuggerPresent devuelve verdadero si se detecta un depurador en el sistema.
func IsDebuggerPresent() bool {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	ret, _, _ := isDebuggerPresent.Call()
	return ret != 0
}
