package acceptor

import (
	"fmt"
	"consensuslib/message"
	"os"
	"encoding/json"
	"net"
	"strconv"
)

type Message = message.Message

type AcceptorRole struct {
	LastPromised Message
	LastAccepted Message
}

func NewAcceptor() AcceptorRole {
	acc := AcceptorRole{
		Message{},
		Message{},
	}
	return acc
}

type AcceptorInterface interface {

	// REQUIRES: a message with the empty/nil/'' string as a value;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessPrepare(msg Message) Message

	// REQUIRES: a message with a value submitted at proposer;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessAccept(msg Message) Message
}

func (acceptor *AcceptorRole) ProcessPrepare(msg Message) Message {
	fmt.Println("[Acceptor] process prepare")
	// no any value had been proposed or n'>n
	// then n' == n and ID' == ID (basically same proposer distributed proposal twice)
	if &acceptor.LastPromised == nil || msg.ID > acceptor.LastPromised.ID {
		acceptor.LastPromised = msg
	} else if acceptor.LastPromised.ID == msg.ID && acceptor.LastPromised.FromProposerID == msg.FromProposerID {
		acceptor.LastPromised = msg
	}
	fmt.Printf("[Acceptor] promised id: %d, val: %s \n", acceptor.LastPromised.ID, acceptor.LastPromised.Value)
	saveIntoFile(acceptor.LastPromised)
	return acceptor.LastPromised
}

func (acceptor *AcceptorRole) ProcessAccept(msg Message) Message {
	fmt.Println("[Acceptor] process accept")
	if &acceptor.LastAccepted == nil {
		if msg.ID == acceptor.LastPromised.ID &&
			msg.FromProposerID == acceptor.LastPromised.FromProposerID {
			acceptor.LastAccepted = msg
		} else if msg.ID > acceptor.LastPromised.ID {
			//acceptor.LastPromised = msg
			acceptor.LastAccepted = msg
		}
	} else {
		if msg.ID == acceptor.LastPromised.ID &&
			acceptor.LastPromised.FromProposerID == msg.FromProposerID {
			acceptor.LastAccepted = msg
		} else if msg.ID > acceptor.LastPromised.ID || msg.ID > acceptor.LastAccepted.ID {
			acceptor.LastAccepted = msg
		}
	}
	fmt.Printf("[Acceptor] accepted id: %d, val: %s \n", acceptor.LastAccepted.ID, acceptor.LastAccepted.Value)
	saveIntoFile(acceptor.LastAccepted)
	return acceptor.LastAccepted

}

// creates a log for acceptor in case of disconnection
func saveIntoFile (msg Message) (err error) {
	addr, errn := net.ResolveTCPAddr("tcp", msg.FromProposerID)
	if errn != nil {
		fmt.Println("[Acceptor] can't resolve own address")
	}
	fmt.Println("[Acceptor] saving message into file")
	var path string
	msgJson, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[Acceptor] errored on marshalling")
		return err
	}
	var f *os.File
	switch msg.Type {
	case message.PREPARE:
		path = strconv.Itoa(addr.Port) + "prepare.json"
	case message.ACCEPT:
		path = strconv.Itoa(addr.Port) + "accept.json"
	}
	if err != nil {
		fmt.Println("[Acceptor] errored on reading path ", err)
	}
	if _, erro := os.Stat(path); os.IsNotExist(erro) {

		f, err = os.Create(path)
		if err != nil {
			fmt.Println("[Acceptor] errored on creating file ", err)
		}

	} else {
		f, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("[Acceptor] errored on opening file ", err)
		}
	}
	defer f.Close()
	_, err = f.Write(msgJson)
	if err != nil {
		fmt.Println("[Acceptor] errored on writing into file ", err)
	}
	f.Close()
	return err
}
