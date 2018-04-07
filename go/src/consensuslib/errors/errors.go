package errors

import (
	"consensuslib/message"
	"fmt"
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
	return fmt.Sprintf("Application server: address already registered [%s]", string(e))
}

type UnknownKeyError string

func (e UnknownKeyError) Error() string {
	return fmt.Sprintf("consensuslib server: unknown key [%s]", string(e))
}

type InvalidLogIndexError string

func (e InvalidLogIndexError) Error() string {
	return fmt.Sprintf("Unable to access the given index in the log.")
}

type ValueForRoundInLogExistsError string

func (e ValueForRoundInLogExistsError) Error() string {
	return fmt.Sprintf("Trying to write a value to an index in the Learner Log that has already been filled")
}

type TimeoutError string

func (e TimeoutError) Error() string {
	return fmt.Sprintf("The function [%s] called timed out.", string(e))
}
