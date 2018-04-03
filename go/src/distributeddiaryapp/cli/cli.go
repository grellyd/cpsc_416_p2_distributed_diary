package cli

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	ALIVE = "alive"
	EXIT  = "exit"
	READ  = "read"
	WRITE = "write"
	HELP  = "help"
	ROUNDS = "rounds"
	PAUSE = "pause"
	CONTINUE = "continue"
)

type Command struct {
	Command string
	Data    *[]string
}

var validCommand = regexp.MustCompile("(alive|read|write ([0-9a-zA-Z ]*)?|help|exit|rounds|pause (prepare|propose|learn|idle)|continue)")

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

pause [prepare|propose|learn|idle]
----------------------------------
- pause the client's execution at the selected stage for the next round until 'continue' is called

continue
--------
- continue the round

Created for:
CPSC 416 Distributed Systems, in the 2017W2 Session at the University of British Columbia (UBC)

Authors: Graham L. Brown (c6y8), Aleksandra Budkina (f1l0b), Larissa Feng (l0j8), Harryson Hu (n5w8), Sharon Yang (l5w8)
`

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
			case 'p':
				when := strings.Split(command[0], " ")[1:]
				return Command{PAUSE, &when}
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
				}
			}
		} else {
			fmt.Println("Command not understood.")
			fmt.Println(helpString)
		}
	}
}

func readFromStdin(reader *bufio.Reader) string {
	in, _ := reader.ReadBytes('\n')
	in = in[:len(in)-1]
	return string(in)
}
