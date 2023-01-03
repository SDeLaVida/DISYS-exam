package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	rep "github.com/SDeLaVida/DISYS-exam/proto"
	"google.golang.org/grpc"
)

type repServer struct {
	rep.ReplicationServer
}

var dict = make(map[string]string, 20)
var amountOfServers = 2
var clients = make([]rep.ReplicationClient, amountOfServers)
var ownPort int
var ownPortStr string
var iAmLeader = false

func updateReplicas(key string, value string) (*rep.AckMessage, error) {
	for _, v := range clients {
		v.Add(context.Background(), key, value)
		returnValue, err := v.Read(context.Background(), key)
		if returnValue != value {
			return &rep.AckMessage{Success: false}, err
		}
	}
	return &rep.AckMessage{Success: true}, nil

}

func main() {
	//Setting portnumber
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort = int(arg1) + 50000
	ownPortStr = strconv.Itoa(int(ownPort))
	log.Println("Starting server on port " + ownPortStr)

	//Creating .log-file for logging output from program, while still printing to the command line
	stringy := fmt.Sprintf("%v_server_output.log", ownPort)
	err := os.Remove(stringy)
	if err != nil {
		log.Println("No previous log file found")
	}
	f, err := os.OpenFile(stringy, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	mw := io.MultiWriter(os.Stdout, f)
	if err != nil {
		fmt.Println("Log does not work")
	}
	defer f.Close()
	log.SetOutput(mw)

	//Listening on own port and creating and setting up server
	list, err := net.Listen("tcp", ":"+ownPortStr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", ownPortStr, err)
	}

	err = initReplicas()
	if err != nil {
		log.Fatalf("Failed to add servers")
	}

	grpcServer := grpc.NewServer()
	rep.RegisterReplicationServer(grpcServer, &repServer{})
	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}

}

func (s *repServer) add(ctx context.Context, key string, value string) *rep.AckMessage {
	/* 	if iAmLeader {
		mess, err := updateReplicas(addMess)
		if err != nil {
			return mess
		}

	} */
	dict[key] = value
	log.Printf("New value is: %s", dict[key])

	if dict[key] != value {
		log.Println("Failure: Did not update value")
		return &rep.AckMessage{Success: false}
	}
	log.Println("Succes: Value updated.")

	return &rep.AckMessage{Success: true}
}

func (s *repServer) read(ctx context.Context, key string) (*rep.ValueMessage, error) {
	return &rep.ValueMessage{
		Value: dict[key],
	}, nil
}

func initReplicas() error {
	for i := 0; i < amountOfServers; i++ {
		if ownPort == 50000+i {
			continue
		}
		portStr := strconv.Itoa(int(50000 + i))

		timeoutConn, err := grpc.Dial(":"+portStr, grpc.WithInsecure())

		if err != nil {
			log.Printf("Dial failed: %v", err)
			return err
		}

		// Defer means: When this function returns, call this method (meaning, one main is done, close connection)
		defer timeoutConn.Close()

		//  Create new Client from generated gRPC code from proto
		c := rep.NewReplicationClient(timeoutConn)
		clients = append(clients, c)

	}
	return nil

}
