package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	ping "github.com/mbjnitu/peer-to-peer/grpc"
	"google.golang.org/grpc"
)

var f *os.File
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
	}

	//Create log.txt if not there
	f, fileerr = os.OpenFile("Log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if fileerr != nil {
		fmt.Println("Error creating log.txt")
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
		p.sendPingToAll()
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id      int32
	lamport int32
	clients map[int32]ping.PingClient
	ctx     context.Context
}

func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	message := req.Message
	recievedLamport := req.Lamport

	fmt.Printf("%v: has a lamport of %v, it received: %v\n", p.id, p.lamport, recievedLamport)
	f.WriteString(fmt.Sprintf("%v: has a lamport of %v, it received: %v\n", p.id, p.lamport, recievedLamport))
	p.lamport = ping.SyncLamport(p.lamport, recievedLamport)
	p.lamport = ping.IncrementLamport(p.lamport) //Receiving a message will increase the Lamport time

	rep := &ping.Reply{Message: message, Lamport: p.lamport}

	return rep, nil
}

func (p *peer) sendPingToAll() {
	for id, client := range p.clients {
		p.lamport = ping.IncrementLamport(p.lamport) //Sending a message will increase the Lamport time
		request := &ping.Request{Message: "hello", Lamport: p.lamport}
		fmt.Printf("%v: send a message with lamport: %v\n", p.id, p.lamport)
		f.WriteString(fmt.Sprintf("%v: send a message with lamport: %v\n", p.id, p.lamport))

		reply, err := client.Ping(p.ctx, request)
		p.lamport = ping.SyncLamport(p.lamport, reply.Lamport) //Receiving a response, that might have a higher Lamport, therefor lets sync.

		if err != nil {
			fmt.Println("something went wrong")
		}

		fmt.Printf("%v: Got reply from id: %v... %v, %v\n", p.id, id, reply.Message, reply.Lamport)
		f.WriteString(fmt.Sprintf("%v: Got reply from id: %v... %v, %v\n", p.id, id, reply.Message, reply.Lamport))
	}
}
