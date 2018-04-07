package cli

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Command is a cli command
type Command struct {
	Command string
	Data    *[]string
}

// Commands
const (
	ALIVE    = "alive"
	EXIT     = "exit"
	READ     = "read"
	WRITE    = "write"
	HELP     = "help"
	ROUNDS   = "rounds"
	BREAK    = "break"
	CONTINUE = "continue"
	STEP     = "step"
	KILL     = "kill"
)

// Breaks
const (
	Prepare = "prepare"
	Propose = "propose"
	Learn   = "learn"
	Idle    = "idle"
	Custom  = "custom"
)

var validCommand = regexp.MustCompile("(alive|read|write ([0-9a-zA-Z ]*)?|help|exit|rounds|(break|kill) (prepare|propose|learn|idle|custom)|continue|step)")

var helpString = `
===========================================
The Chamber of Secrets: A Distributed Diary
===========================================
Valid Commands:
   
alive
-----
- Report if this client is connected to the server

exit
----
- exit the program

help
----
- display this text

read
----
- read the current log value of the application

write [a-zA-Z0-9 ]?
-------------------
- write to the log a string consisiting of one or more lower and upper case letters, 0-9, and spaces.

rounds
-------
- produce the round results from the paxostracker

break [prepare|propose|learn|idle|custom]
----------------------------------
- break the client's execution at the selected stage for the next round until 'continue' is called

kill [prepare|propose|learn|idle|custom]
----------------------------------
- kill the client's execution at the selected stage. Exits roughly with os.Exit(1).

continue
--------
- continue the round

step
----
- step one stage further

Created for:
CPSC 416 Distributed Systems, in the 2017W2 Session at the University of British Columbia (UBC)

Authors: Graham L. Brown (c6y8), Aleksandra Budkina (f1l0b), Larissa Feng (l0j8), Harryson Hu (n5w8), Sharon Yang (l5w8)
`

// Run the cli
func Run() (cmd Command) {
	for {
		fmt.Printf("[DD]:")
		reader := bufio.NewReader(os.Stdin)
		inputString := readFromStdin(reader)
		command := validCommand.FindStringSubmatch(inputString)
		if command != nil && len(command) > 0 {
			switch command[0][0] {
			case 'w':
				// split string for written string
				writeArgs := strings.Split(command[0], " ")[1:]
				return Command{WRITE, &writeArgs}
			case 'b':
				when := strings.Split(command[0], " ")[1:]
				return Command{BREAK, &when}
			case 'k':
				when := strings.Split(command[0], " ")[1:]
				return Command{KILL, &when}
			default:
				switch command[0] {
				case ALIVE:
					return Command{ALIVE, nil}
				case READ:
					return Command{READ, nil}
				case EXIT:
					return Command{EXIT, nil}
				case HELP:
					fmt.Println(helpString)
				case ROUNDS:
					return Command{ROUNDS, nil}
				case CONTINUE:
					return Command{CONTINUE, nil}
				case STEP:
					return Command{STEP, nil}
				default:
					fmt.Println("Command not understood.")
					fmt.Println("Type 'help' for command information.")
				}
			}
		} else {
			fmt.Println("Command not understood.")
			fmt.Println("Type 'help' for command information.")
		}
	}
}

func readFromStdin(reader *bufio.Reader) string {
	in, _ := reader.ReadBytes('\n')
	in = in[:len(in)-1]
	return string(in)
}
