package errors

import (
	"fmt"
	"consensuslib/message"
)

type InvalidMessageTypeError message.Message

func (e InvalidMessageTypeError) Error() string {
	return fmt.Sprintf("This is an invalid message type. Message type should only be PREPARE, ACCEPT, CONSENSUS")
}

type NeighbourConnectionError string

func (e NeighbourConnectionError) Error() string {
	return fmt.Sprintf("Unable to open RPC connection with a new neighbour that connected to PN")
}

type AddressAlreadyRegisteredError string

func (e AddressAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("BlockArt server: address already registered [%s]", string(e))
}

type UnknownKeyError string

func (e UnknownKeyError) Error() string {
	return fmt.Sprintf("consensuslib server: unknown key [%s]", string(e))
}
