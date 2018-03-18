package consensuslib

import "fmt"

type InvalidMessageTypeError Message

func (e InvalidMessageTypeError) Error() string {
	return fmt.Sprintf("This is an invalid message type. Message type should only be PREPARE, ACCEPT, CONSENSUS")
}

type NeighbourConnectionError string

func (e NeighbourConnectionError) Error() string {
	return fmt.Sprintf("Unable to open RPC connection with a new neighbour that connected to PN")
}
