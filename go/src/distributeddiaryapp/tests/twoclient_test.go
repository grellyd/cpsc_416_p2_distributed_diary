package tests

import (
	"distributeddiaryapp/tests/util"
	"testing"
	"time"
)

type TestTwoData struct {
	DataC0 string
	DataC1 string
}

func TwoTests() []TestTwoData {
	return []TestTwoData{
		{
			DataC0: "beep",
			DataC1: "boop bop",
		},
		{
			DataC0: "Voldie Sucks!",
			DataC1: "No, Voldemort Rocks",
		},
	}
}

func TestTwoReadOneWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localPort := "0"
	util.SetupServer(serverAddr)
	for _, test := range TwoTests() {
		client0, err := util.SetupClient(serverAddr, localPort)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadOneWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := util.SetupClient(serverAddr, localPort)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// C0 Writes
		err = client0.Write(test.DataC0)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadOneWrite(%v)\" produced err: %v", test, err)
		}
		// C0 Reads
		value, err := client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see it's own value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// Can C1 see C0's value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, test.DataC0)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestTwoReadTwoWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localPort := "127.0.0.1:0"
	util.SetupServer(serverAddr)
	for _, test := range TwoTests() {
		client0, err := util.SetupClient(serverAddr, localPort)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := util.SetupClient(serverAddr, localPort)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// C0 Writes
		err = client0.Write(test.DataC0)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		// C0 Reads
		value, err := client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see it's own value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C1 see C0's value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Writes
		err = client1.Write(test.DataC1)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C1 see the combined log?
		combinedData := test.DataC0 + " " + test.DataC1
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, combinedData)
		}

		// C0 Reads
		value, err = client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see the combined log?
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, combinedData)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
