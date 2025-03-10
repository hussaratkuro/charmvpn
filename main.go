package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

type Action string

const (
	Connect    Action = "Connect to VPN"
	Disconnect Action = "Disconnect from VPN"
	ListVPNs   Action = "List available VPNs"
	Status     Action = "Show VPN status"
	AddVPN     Action = "Add VPN"
	RemoveVPN  Action = "Remove VPN"
	Exit       Action = "Exit"
)

func executeCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %s\n%s", err, output)
	}
	return string(output)
}

func listVPNs() string {
	// Get all connections with their types
	output := executeCommand("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	var vpns []string
	for _, line := range lines {
		if strings.Contains(line, ":vpn") {
			vpn := strings.Split(line, ":")[0]
			vpns = append(vpns, vpn)
		}
	}
	
	if len(vpns) == 0 {
		return "No VPN connections found"
	}
	
	var result strings.Builder
	result.WriteString("Available VPN connections:\n")
	for i, vpn := range vpns {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, vpn))
	}
	
	return result.String()
}

func connectVPN(vpnName string) string {
	return executeCommand("nmcli", "connection", "up", vpnName)
}

func getActiveVPNs() []string {
	output := executeCommand("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show", "--active")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	var vpns []string
	for _, line := range lines {
		if strings.Contains(line, ":vpn") {
			vpn := strings.Split(line, ":")[0]
			vpns = append(vpns, vpn)
		}
	}
	return vpns
}

func disconnectVPN() string {
	vpns := getActiveVPNs()
	if len(vpns) == 0 {
		return "No active VPN connections found"
	}
	
	var result strings.Builder
	for _, vpn := range vpns {
		output := executeCommand("nmcli", "connection", "down", vpn)
		result.WriteString(fmt.Sprintf("Disconnecting %s: %s\n", vpn, output))
	}
	return result.String()
}

func vpnStatus() string {
	activeVpns := getActiveVPNs()
	if len(activeVpns) == 0 {
		return "No active VPN connections"
	}
	
	var result strings.Builder
	result.WriteString("Active VPN connections:\n")
	
	for _, vpn := range activeVpns {
		details := executeCommand("nmcli", "connection", "show", vpn)
		result.WriteString(fmt.Sprintf("--- %s ---\n%s\n", vpn, details))
	}
	
	return result.String()
}

func addVPN(vpnFile string) string {
	return executeCommand("nmcli", "connection", "import", "type", "openvpn", "file", vpnFile)
}

func removeVPN(vpnName string) string {
	return executeCommand("nmcli", "connection", "delete", vpnName)
}

func main() {
	for {
		var action Action
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[Action]().
					Title("Choose an action").
					Value(&action).
					Options(
						huh.NewOption[Action](string(Connect), Connect),
						huh.NewOption[Action](string(Disconnect), Disconnect),
						huh.NewOption[Action](string(ListVPNs), ListVPNs),
						huh.NewOption[Action](string(Status), Status),
						huh.NewOption[Action](string(AddVPN), AddVPN),
						huh.NewOption[Action](string(RemoveVPN), RemoveVPN),
						huh.NewOption[Action](string(Exit), Exit),
					),
			),
		)
		
		if err := form.Run(); err != nil {
			fmt.Println("Error:", err)
			return
		}
		
		switch action {
		case Connect:
			var vpnName string
			vpnForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter VPN name").
						Value(&vpnName),
				),
			)
			if err := vpnForm.Run(); err == nil {
				fmt.Println(connectVPN(strings.TrimSpace(vpnName)))
			}
			
		case Disconnect:
			fmt.Println(disconnectVPN())
			
		case ListVPNs:
			fmt.Println(listVPNs())
			
		case Status:
			fmt.Println("VPN Status:")
			fmt.Println(vpnStatus())
			
		case AddVPN:
			var vpnFile string
			vpnFileForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter path to .ovpn file").
						Value(&vpnFile),
				),
			)
			if err := vpnFileForm.Run(); err == nil {
				fmt.Println(addVPN(strings.TrimSpace(vpnFile)))
			}
			
		case RemoveVPN:
			var vpnName string
			vpnRemoveForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter VPN name to remove").
						Value(&vpnName),
				),
			)
			if err := vpnRemoveForm.Run(); err == nil {
				fmt.Println(removeVPN(strings.TrimSpace(vpnName)))
			}
			
		case Exit:
			return
		}
	}
}