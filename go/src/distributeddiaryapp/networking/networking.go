package networking

import (
	"fmt"
	"strings"
	"os/exec"
)

// GetOutboundIP Returns a machine's public (outbound) Azure IP address e.g. "270.0.21.1"
func GetOutboundIP() (ipString string, err error) {
	out, err := exec.Command("curl", "-s", "http://checkip.amazonaws.com").Output()

	if err != nil {
		fmt.Println(err)
	}

	result := fmt.Sprintf("%s", out)
	return strings.TrimSpace(result), nil
}

