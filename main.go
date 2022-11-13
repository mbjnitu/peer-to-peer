package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	ping "github.com/mbjnitu/peer-to-peer/grpc"
	"google.golang.org/grpc"
)

var logFile, CRITICAL_FILE *os.File
var fileerr error

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:      ownPort,
		lamport: 0,
		clients: make(map[int32]ping.PingClient),
		ctx:     ctx,
		state:   "none",
	}

	//Create log.txt if not there
	logFile, fileerr = os.OpenFile("Log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if fileerr != nil {
		fmt.Println("Error opening log.txt")
	}
	CRITICAL_FILE, fileerr = os.OpenFile("CRITICAL_SECTION.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if fileerr != nil {
		fmt.Println("Error opening log.txt")
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := ping.NewPingClient(conn)
		p.clients[port] = c
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.sendPingToAll(scanner.Text())
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id      int32
	lamport int32
	clients map[int32]ping.PingClient
	ctx     context.Context
	state   string
}

func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	message := req.Message
	recievedLamport := req.Lamport
	var recievedId int32 = 0
	var response string = "yes"

	fmt.Printf("%v: has a lamport of %v, it received: %v (%v)\n", p.id, p.lamport, message, recievedLamport)
	logFile.WriteString(fmt.Sprintf("%v: has a lamport of %v, it received: %v (%v)\n", p.id, p.lamport, message, recievedLamport))

	if message == "may i enter" {
		// Do i want to enter aswell?
		random := rand.Intn(100)
		rand.NewSource(time.Now().UnixNano())
		if p.state == "in" { //Im aleready using the critical section, like wtf are you thinking?..
			response = "no"
		} else if random > 65 {
			if (p.lamport < req.Lamport) || (p.lamport == req.Lamport && p.id > recievedId) {
				response = "no"
				go p.waitAndRequest()
			}
		}
	}

	p.lamport = ping.SyncLamport(p.lamport, recievedLamport)
	p.lamport = ping.IncrementLamport(p.lamport) //Receiving a message will increase the Lamport time

	rep := &ping.Reply{Message: response, Lamport: p.lamport}

	return rep, nil
}

func (p *peer) waitAndRequest() {
	time.Sleep(2000 * time.Millisecond) // Since i also wanted to enter the critical section, and my lamport/id-combo was superior, ill now make a request myself.
	p.sendPingToAll("may i enter")
}

func (p *peer) sendPingToAll(message string) {
	amountOfYes := 0
	for id, client := range p.clients {
		p.lamport = ping.IncrementLamport(p.lamport) //Sending a message will increase the Lamport time
		request := &ping.Request{Message: message, Lamport: p.lamport}
		fmt.Printf("%v: send message '%v' with lamport: %v\n", p.id, message, p.lamport)
		logFile.WriteString(fmt.Sprintf("%v: send message '%v' with lamport: %v\n", p.id, message, p.lamport))

		reply, err := client.Ping(p.ctx, request)
		p.lamport = ping.SyncLamport(p.lamport, reply.Lamport) //Receiving a response, that might have a higher Lamport, therefor lets sync.

		if err != nil {
			fmt.Println("something went wrong")
		}

		if reply.Message == "yes" {
			amountOfYes++
		}

		fmt.Printf("%v: Got reply from id: %v... %v, %v\n", p.id, id, reply.Message, reply.Lamport)
		logFile.WriteString(fmt.Sprintf("%v: Got reply from id: %v... %v, %v\n", p.id, id, reply.Message, reply.Lamport))
	}

	if amountOfYes == (3)-1 {
		p.writeInCriticalSection()
	}
}

func (p *peer) writeInCriticalSection() {
	p.state = "in"
	CRITICAL_FILE.WriteString(fmt.Sprintf("%v: Accessed the Critical section. Lamport: %v\n", p.id, p.lamport))
	logFile.WriteString(fmt.Sprintf("%v: Accessed the Critical section. Lamport: %v\n", p.id, p.lamport))
	time.Sleep(8000 * time.Millisecond)
	p.state = "none"
}
