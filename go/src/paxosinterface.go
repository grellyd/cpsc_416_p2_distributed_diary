package paxosinterface

// Functions for a client to interact with the paxos network
// The paxos net will maintain the game state between the two teams
type PaxosInterface interface {
// Add functions that a client will call
  MakeMove()
  GetBoard()
  
  // This gets broadcast to all paxos nodes (implemented on paxos net side)
  IsNodeLeader()
}
