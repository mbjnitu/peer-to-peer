# peer-to-peer
Solution for Dist Sys, Mandatory Handin 4 - Distributed Mutual Exclusion Assignment

# How to run
The program runs by connecting 3 peers to eatch other.
This is done by running the main.go file from 3 different CMD / consoles.
Each should be run with the following commands
1. "go run main.go 0"
2. "go run main.go 1"
3. "go run main.go 2"

Now the user can choose which of the other clients that should request to enter the critical section.
This is done by typing "may i enter" in one of the processes' console.
Now the others clients will decide if they to wish to enter (this is done randomly).
The program will now simulate a series of requests, where the requesting process with the lowest lamport time wins.

When this is done, the user can once again write "may i enter" in one of them.
