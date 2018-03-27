package networking

import (
	"fmt"
	"net"
	"strings"
)

func GetOutboundIP() (ipString string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Outbound IP couldn't be fetched")
		return "", err
	}

	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	index := strings.LastIndex(localAddr, ":")
	return localAddr[0:index], nil
}
