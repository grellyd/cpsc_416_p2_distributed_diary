package util

import (
	"consensuslib"
	"time"
)

const (
	HEARTBEAT_INTERVAL = 1*time.Millisecond
)

func SetupClient(serverAddr string, localAddr string) (client *consensuslib.Client, err error) {
	client, err = consensuslib.NewClient(localAddr, HEARTBEAT_INTERVAL)
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
