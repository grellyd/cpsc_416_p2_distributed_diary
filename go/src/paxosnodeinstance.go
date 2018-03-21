package paxosnode

import (
	"consensuslib/paxosnode/paxosnodeinterface"
	"fmt"
	"net/rpc"
	"net"
	"time"
	"consensuslib/message"
)

type Message = message.Message
