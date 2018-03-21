package tests

import (
	"time"
	"testing"
	"consensuslib"
)

func TestTwoClientsReadWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localAddr := "127.0.0.1:0"
	var tests = []struct {
		DataC0       string
		DataC1       string
	}{
		{
			DataC0: "testing more",
			DataC1: "testing",
		},
		{
			DataC0: "Voldie Sucks",
			DataC1: "No, Voldemort Rocks",
		},
	}
	for _, test := range tests {
		client0, err := setupClient(serverAddr, localAddr, 1*time.Millisecond)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := setupClient(serverAddr, localAddr, 1*time.Millisecond)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}

		// C0 Writes
		err = client0.Write(test.DataC0)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}
		// C0 Reads
		value, err := client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see it's own value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}
		
		// Can C1 see C0's value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Writes
		err = client1.Write(test.DataC1)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}
		
		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}
		
		// Can C1 see the combined log?
		combinedData := test.DataC0 + test.DataC1 
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, combinedData)
		}
		
		// C0 Reads
		value, err = client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoClientsReadWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see the combined log?
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, combinedData)
		}
	}
}

func setupClient(serverAddr string, localAddr string, heartbeatRate time.Duration) (client *consensuslib.Client, err error) {
	client, err = consensuslib.NewClient(localAddr, heartbeatRate)
	if err != nil {
		return nil, err
	}
	err = client.Connect(serverAddr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
