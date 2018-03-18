package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Nserver int

type User struct {
	Address   string
	Heartbeat int64
}

type AllUsers struct {
	sync.RWMutex
	all map[string]*User
}

/*type ClientSettings struct {
	Leader bool
	Teammates []string
	Enemies []string
}*/

type HeartBeat uint32

var (
	heartBeat HeartBeat   = 2
	errLog    *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog    *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	allUsers  AllUsers    = AllUsers{all: make(map[string]*User)}
)

type AddressAlreadyRegisteredError string

func (e AddressAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("BlockArt server: address already registered [%s]", string(e))
}

// Registers a client with the server
func (s *Nserver) Register(addr string, res *[]string) error {
	allUsers.Lock()
	defer allUsers.Unlock()

	if _, exists := allUsers.all[addr]; exists {
		return AddressAlreadyRegisteredError(addr)
	}
	allUsers.all[addr] = &User{
		addr,
		time.Now().UnixNano(),
	}

	go monitor(addr, time.Duration(heartBeat)*time.Second)

	neighbourAddresses := make([]string, 0)

	for _, val := range allUsers.all {
		if addr == val.Address {
			continue
		}
		neighbourAddresses = append(neighbourAddresses, val.Address)
	}
	*res = neighbourAddresses

	outLog.Printf("Got Register from %s\n", addr)

	return nil

}

// from proj1 server.go implementation by Ivan Beschastnikh
func (s *Nserver) HeartBeat(addr string, _ignored *bool) error {
	allUsers.Lock()
	defer allUsers.Unlock()

	if _, ok := allUsers.all[addr]; !ok {
		return errors.New("Server: unknown key")
	}

	allUsers.all[addr].Heartbeat = time.Now().UnixNano()

	return nil
}

func (s *Nserver) CheckAlive(addr string, alive *bool) error {
	*alive = true
	return nil
}

// from proj1 server.go implementation by Ivan Beschastnikh
func monitor(k string, heartBeatInterval time.Duration) {
	for {
		allUsers.Lock()
		if time.Now().UnixNano()-allUsers.all[k].Heartbeat > int64(heartBeatInterval) {
			outLog.Printf("%s timed out\n", allUsers.all[k].Address)
			delete(allUsers.all, k)
			allUsers.Unlock()
			return
		}
		outLog.Printf("%s is alive\n", allUsers.all[k].Address)
		allUsers.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func main() {
	// register entity required to recieve RPC calls
	nserver := new(Nserver)
	server := rpc.NewServer()
	server.Register(nserver)
	servAddr := "127.0.0.1:12345"
	l, e := net.Listen("tcp", servAddr)
	checkError(e, "Connection error")
	fmt.Println("Server started at ", servAddr)

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}

}

func checkError(e error, m string) {
	if e != nil {
		errLog.Fatalf("%s, err = %s\n", m, e.Error())
	}
}
