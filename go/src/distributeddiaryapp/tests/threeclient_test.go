package tests

import (
	"distributeddiaryapp/tests/util"
	"testing"
)

type TestThreeData struct {
	DataC0 string
	DataC1 string
	DataC2 string
}

func ThreeTests() []TestThreeData {
	return []TestThreeData{
		{
			DataC0: "beep",
			DataC1: "boop bop",
			DataC2: "further testing isn't needed",
		},
		{
			DataC0: "Voldie Sucks!",
			DataC1: "No, Voldemort Rocks;",
			DataC2: "Avada Kedavra!",
		},
	}
}

func TestThreeReadOneWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localAddr := "127.0.0.1:0"
	util.SetupServer(serverAddr)
	for _, test := range ThreeTests() {
		client0, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}
		client2, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// C0 Writes
		err = client0.Write(test.DataC0)
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}
		// C0 Reads
		value, err := client0.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// Can C0 see it's own value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 0 does not match written data '%s'", value, test.DataC0)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// Can C1 see C0's value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 1 does not match written data '%s'", value, test.DataC0)
		}

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestThreeReadOneWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see C0's value?
		if value != test.DataC0 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, test.DataC0)
		}
	}
}

func TestThreeReadTwoWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localAddr := "127.0.0.1:0"
	util.SetupServer(serverAddr)
	for _, test := range ThreeTests() {
		client0, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		client2, err := util.SetupClient(serverAddr, localAddr)
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

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see C0's value?
		if value != test.DataC2 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, test.DataC0)
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

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see the combined log?
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, combinedData)
		}
	}
}

func TestThreeReadThreeWrite(t *testing.T) {
	serverAddr := "127.0.0.1:12345"
	localAddr := "127.0.0.1:0"
	util.SetupServer(serverAddr)
	for _, test := range ThreeTests() {
		client0, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		client1, err := util.SetupClient(serverAddr, localAddr)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}
		client2, err := util.SetupClient(serverAddr, localAddr)
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

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see C0's value?
		if value != test.DataC2 {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, test.DataC0)
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

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see the combined log?
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, combinedData)
		}

		// C2 Writes
		err = client2.Write(test.DataC2)
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// C1 Reads
		value, err = client1.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		combinedData = test.DataC0 + " " + test.DataC1 + " " + test.DataC2

		// Can C1 see the combined log?
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

		// C2 Reads
		value, err = client2.Read()
		if err != nil {
			t.Errorf("Bad Exit: \"TestTwoReadTwoWrite(%v)\" produced err: %v", test, err)
		}

		// Can C2 see the combined log?
		if value != combinedData {
			t.Errorf("Bad Exit: Read Data '%s' for Client 2 does not match written data '%s'", value, combinedData)
		}
	}
}
