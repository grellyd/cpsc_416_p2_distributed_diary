package tests

import (
	"distributeddiaryapp/tests/util"
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
	util.SetupServer(serverAddr)
	for _, test := range tests {
		client, err := util.SetupClient(serverAddr, localAddr)
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
		time.Sleep(5 * time.Millisecond)
	}
}
