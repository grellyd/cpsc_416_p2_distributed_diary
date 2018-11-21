============================================
The Chamber of Secrets: A Distributed Diary
============================================

Created for:
CPSC 416 Distributed Systems, in the 2017W2 Session at the University of
British Columbia (UBC) to demonstrate the PAXOS Consensus Algorithm

https://www.cs.ubc.ca/~bestchai/teaching/cs416_2017w2/project2/index.html

Authors: Graham L. Brown (c6y8), Aleksandra Budkina (f1l0b), Larissa Feng (l0j8), Harryson Hu (n5w8), Sharon Yang (l5w8)

The application can be run both on a local machine and on Azure with different VMs.

Let PORT be some valid free port number.

To run the app locally:
— Start server from go/src: go run distributeddiaryserver/server.go 12345 local
- Start app from go/src: go run distributeddiaryapp/app.go 127.0.0.1:12345 PORT --local

This will run server and apps at 127.0.0.1:PORT

Let ADDRESS be the outgoing IP for the server VM.

To run the app globally:
— Start server from src: go run distributeddiaryserver/server.go 12345
- Start app from src: go run distributeddiaryapp/app.go ADDRESS:12345 PORT

This will run apps on machine's outbound IP on port PORT

The performance logs are stored under src/logs
To view the performance at real time add “--debug” in the end of the command that runs the app.

Steps to reproduce the failure cases of section 2.4 of the final report are stored at failure_case_playbook.txt
