package consensuslib

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

type Server struct {
	rpcServer *rpc.Server
	listener net.Listener
}

type User struct {
	Address   string
	Heartbeat int64
}

type AllUsers struct {
	sync.RWMutex
	all map[string]*User
}

type HeartBeat uint32

var (
	heartBeat HeartBeat   = 2
	errLog    *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog    *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	allUsers  AllUsers    = AllUsers{all: make(map[string]*User)}
)

// Creates a new server ready to register paxosnodes
// TODO: inject logger
func NewServer(addr string) (server *Server, err error) {
	server = &Server{
		rpcServer: rpc.NewServer(),
	}
	server.rpcServer.Register(server)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// TODO: enhance error
		return nil, err
	}
	server.listener = listener
	fmt.Println("Server started at ", addr)
	return server, nil
}

func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("unable to accept connection: %s", err)
		}
		go s.rpcServer.ServeConn(conn)
	}
}


// Registers a client with the server
func (s *Server) Register(addr string, res *[]string) error {
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
func (s *Server) HeartBeat(addr string, _ignored *bool) error {
	allUsers.Lock()
	defer allUsers.Unlock()

	if _, ok := allUsers.all[addr]; !ok {
		return errors.New("Server: unknown key")
	}

	allUsers.all[addr].Heartbeat = time.Now().UnixNano()

	return nil
}

func (s *Server) CheckAlive(addr string, alive *bool) error {
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

// TODO: Not fail fatal. Pass up to caller
func checkError(e error, m string) {
	if e != nil {
		errLog.Fatalf("%s, err = %s\n", m, e.Error())
	}
}
