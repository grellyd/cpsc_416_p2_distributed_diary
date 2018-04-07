package consensuslib

import (
	"consensuslib/errors"
	"filelogger/singletonlogger"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Server is our server
type Server struct {
	rpcServer *rpc.Server
	listener  net.Listener
}

// User represents a connected client
type User struct {
	Address   string
	Heartbeat int64
}

// AllUsers is the collection of all our users
type AllUsers struct {
	sync.RWMutex
	all map[string]*User
}

// HeartBeat is our heartbeat rate
type HeartBeat uint32

var (
	heartBeat HeartBeat = 2
	allUsers            = AllUsers{all: make(map[string]*User)}
)

// NewServer creates a new server ready to register paxosnodes
func NewServer(addr string) (server *Server, err error) {
	server = &Server{
		rpcServer: rpc.NewServer(),
	}
	server.rpcServer.Register(server)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("unable to create a listener on the server addres: %s", err)
	}
	server.listener = listener
	singletonlogger.Info("Server started at " + listener.Addr().String())
	return server, nil
}

// Serve for clients
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("[ConsensusLib/serv] Unable to accept connection: %s", err)
		}
		singletonlogger.Debug(fmt.Sprintf("[ConsensusLib/serv] Serving %s\n", s.listener.Addr().String()))
		go s.rpcServer.ServeConn(conn)
	}
}

// Register a client with the server
func (s *Server) Register(addr string, res *[]string) error {
	allUsers.Lock()
	defer allUsers.Unlock()

	if _, exists := allUsers.all[addr]; exists {
		return errors.AddressAlreadyRegisteredError(addr)
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

	singletonlogger.Info(fmt.Sprintf("Got Register from %s", addr))

	return nil

}

// HeartBeat from proj1 server.go implementation by Ivan Beschastnikh
func (s *Server) HeartBeat(addr string, _ignored *bool) error {
	allUsers.Lock()
	defer allUsers.Unlock()

	if _, ok := allUsers.all[addr]; !ok {
		// TODO: check right chanage
		return errors.UnknownKeyError("")
	}

	allUsers.all[addr].Heartbeat = time.Now().UnixNano()

	return nil
}

// CheckAlive says if the server is alive
func (s *Server) CheckAlive(addr string, alive *bool) error {
	*alive = true
	return nil
}

// from proj1 server.go implementation by Ivan Beschastnikh, adapted by Graham Brown
func monitor(k string, heartBeatInterval time.Duration) {
	for {
		allUsers.Lock()
		if time.Now().UnixNano()-allUsers.all[k].Heartbeat > int64(heartBeatInterval) {
			singletonlogger.Info(fmt.Sprintf("%s timed out", allUsers.all[k].Address))
			delete(allUsers.all, k)
			allUsers.Unlock()
			return
		}
		singletonlogger.Info(fmt.Sprintf("%s is alive", allUsers.all[k].Address))
		allUsers.Unlock()
		time.Sleep(heartBeatInterval)
	}
}
