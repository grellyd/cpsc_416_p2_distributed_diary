package util

import (
	"consensuslib"
	"strconv"
	"time"
)

const (
	HEARTBEAT_INTERVAL = 1 * time.Millisecond
)

func SetupClient(serverAddr string, localPort string) (client *consensuslib.Client, err error) {
	intPort, err := strconv.Atoi(localPort)
	client, err = consensuslib.NewClient(intPort, HEARTBEAT_INTERVAL)
	if err != nil {
		return nil, err
	}
	err = client.Connect(serverAddr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func SetupServer(serverAddr string) (err error) {
	server, err := consensuslib.NewServer(serverAddr)
	if err != nil {
		return err
	}
	go server.Serve()
	return nil
}
