package tests

import (
	"consensuslib"
	"testing"
	"time"
)

func TestSingleClientReadWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localAddr := "127.0.0.1:0"
	var tests = []struct {
		Data string
	}{
		{
			Data: "testing",
		},
		{
			Data: "Voldemort Rocks",
		},
	}
	server, err := consensuslib.NewServer(serverAddr)
	if err != nil {
		t.Errorf("Bad Exit: \"TestSingleClientReadWrite()\" produced err: %v", err)
	}
	go server.Serve()
	for _, test := range tests {
		client, err := consensuslib.NewClient(localAddr, 1*time.Millisecond)
		if err != nil {
			t.Errorf("Bad Exit: \"TestSingleClientReadWrite(%v)\" produced err: %v", test, err)
		}
		err = client.Connect(serverAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestSingleClientReadWrite(%v)\" produced err: %v", test, err)
		}
		err = client.Write(test.Data)
		if err != nil {
			t.Errorf("Bad Exit: \"TestSingleClientReadWrite(%v)\" produced err: %v", test, err)
		}
		value, err := client.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestSingleClientReadWrite(%v)\" produced err: %v", test, err)
		}
		if value != test.Data {
			t.Errorf("Bad Exit: Read Data '%s' does not match written data '%s'", value, test.Data)
		}
	}
}
