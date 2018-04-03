package acceptor

import (
	"filelogger/singletonlogger"
	"consensuslib/message"
	"encoding/json"
	"fmt"
	//"net"
	"os"
	//"strconv"
	"io/ioutil"
	"math/rand"
	"time"
)

type Message = message.Message

type AcceptorRole struct {
	ID 			 string
	LastPromised Message
	LastAccepted Message
}

func NewAcceptor() AcceptorRole {
	id := generateAcceptorID(6)
	acc := AcceptorRole{
		id,
		Message{},
		Message{},
	}
	singletonlogger.Debug(fmt.Sprintf("[Acceptor] %v", acc.ID))
	return acc
}

type AcceptorInterface interface {

	// REQUIRES: a message with the empty/nil/'' string as a value;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessPrepare(msg Message) Message

	// REQUIRES: a message with a value submitted at proposer;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessAccept(msg Message) Message

	// EFFECTS: returns the last accepted message if any
	RestoreFromBackup(port string)
}

func (acceptor *AcceptorRole) ProcessPrepare(msg Message, roundNum int) Message {
	singletonlogger.Debug(fmt.Sprintf("[Acceptor] process prepare for round %v", roundNum))
	// no any value had been proposed or n'>n
	// then n' == n and ID' == ID (basically same proposer distributed proposal twice)
	if &acceptor.LastPromised == nil ||
		(msg.ID > acceptor.LastPromised.ID && roundNum >= acceptor.LastPromised.RoundNum) {
		acceptor.LastPromised = msg
	} else if acceptor.LastPromised.ID > msg.ID &&
		//acceptor.LastPromised.FromProposerID == msg.FromProposerID &&
			acceptor.LastPromised.RoundNum == roundNum {
		acceptor.LastPromised = msg
	}
	singletonlogger.Debug(fmt.Sprintf("[Acceptor] promised id: %d, val: %s, round: %d \n", acceptor.LastPromised.ID, acceptor.LastPromised.Value, roundNum))
	acceptor.saveIntoFile(acceptor.LastPromised)
	return acceptor.LastPromised
}

func (acceptor *AcceptorRole) ProcessAccept(msg Message, roundNum int) Message {
	singletonlogger.Debug("[Acceptor] process accept")
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
			acceptor.LastPromised.FromProposerID == msg.FromProposerID &&
				acceptor.LastPromised.RoundNum == roundNum {
			acceptor.LastAccepted = msg
		} else if (msg.ID > acceptor.LastPromised.ID && acceptor.LastPromised.RoundNum >= roundNum) ||
			(msg.ID > acceptor.LastAccepted.ID && acceptor.LastAccepted.RoundNum >= roundNum) {
			acceptor.LastAccepted = msg
		}
	}
	singletonlogger.Debug(fmt.Sprintf("[Acceptor] accepted id: %d, val: %s, round: %d \n", acceptor.LastAccepted.ID, acceptor.LastAccepted.Value, roundNum))
	acceptor.saveIntoFile(acceptor.LastAccepted)
	return acceptor.LastAccepted

}

// TODO: since we're testing on the same machine use a port as a reference point
// TODO: in the last version this method will have nothing, because we'll be running
// code on different machines
func (acceptor *AcceptorRole) RestoreFromBackup() {
	singletonlogger.Debug("[Acceptor] restoring from backup")
	path := "temp1/"+acceptor.ID + "prepare.json"
	f, err := os.Open(path)
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] no such file exist, no messages were promised %v", err))
		return
	}
	buf, err := ioutil.ReadAll(f)
	err = json.Unmarshal(buf, &acceptor.LastPromised)
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] error on unmarshalling promise %v", err))
	}
	f.Close()
	path = "temp1/"+acceptor.ID + "accept.json"
	f, err = os.Open(path)
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] no such file exist, no messages were accepted %v", err))
		return
	}
	buf, err = ioutil.ReadAll(f)
	err = json.Unmarshal(buf, &acceptor.LastAccepted)
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] error on unmarshalling accept %v", err))
	}
}

// creates a log for acceptor in case of disconnection
func (a *AcceptorRole)saveIntoFile(msg Message) (err error) {
	/*addr, errn := net.ResolveTCPAddr("tcp", msg.FromProposerID)
	if errn != nil {
		singletonlogger.Debug("[Acceptor] can't resolve own address")
	}*/
	singletonlogger.Debug("[Acceptor] saving message into file")
	var path string
	msgJson, err := json.Marshal(msg)
	if err != nil {
		singletonlogger.Debug("[Acceptor] errored on marshalling")
		return err
	}
	var f *os.File
	switch msg.Type {
	case message.PREPARE:
		path = "temp1/"+ a.ID + "prepare.json"
		singletonlogger.Debug("[Acceptor] saved PREPARE to file")
	case message.ACCEPT:
		path = "temp1/"+ a.ID + "accept.json"
		singletonlogger.Debug("[Acceptor] saved ACCEPT to file")
	}
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] errored on reading path %v", err))
	}
	if _, erro := os.Stat(path); os.IsNotExist(erro) {
		os.MkdirAll("temp1/", os.ModePerm);
		f, err = os.Create(path)
		if err != nil {
			singletonlogger.Debug(fmt.Sprintf("[Acceptor] errored on creating file %v", err))
		}

	} else {
		f, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			singletonlogger.Debug(fmt.Sprintf("[Acceptor] errored on opening file %v", err))
		}
	}
	//defer f.Close()
	_, err = f.Write(msgJson)
	if err != nil {
		singletonlogger.Debug(fmt.Sprintf("[Acceptor] errored on writing into file %v", err))
	}
	f.Close()
	//singletonlogger.Debug("[Acceptor] saved message into file")
	return err
}

func generateAcceptorID(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
