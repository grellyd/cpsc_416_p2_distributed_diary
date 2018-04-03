package networking

import (
	"fmt"
	"os/exec"
)

// GetOutboundIP Returns a machine's public (outbound) IP address e.g. "270.0.21.1".
func GetOutboundIP() (ipString string, err error) {
	out, err := exec.Command("curl -s http://checkip.amazonaws.com || printf \"0.0.0.0\"").Output()

	if err != nil {
		fmt.Println(err)
	}

	return fmt.Sprintf("%s", out), nil
}
