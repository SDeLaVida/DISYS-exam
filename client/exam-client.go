package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	rep "github.com/SDeLaVida/DISYS-exam/proto"
	"google.golang.org/grpc"
)

var (
	totalPorts = 2
	id         int
	mainClient rep.TemplateClient()
	reader     = bufio.NewReader(os.Stdin)
)

func main() {
	//Loading id and total amount of ports to connect to
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	id = int(arg1)

	//Creating .log-file for logging output from program, while still printing to the command line
	stringy := fmt.Sprintf("%v_client_output.log", id)
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

	//Creating connection to main server
	for i := 0; i < totalPorts; i++ {
		// Create a virtual RPC Client Connection on port 9080 + i
		portStr := strconv.Itoa(50000 + i)
		conn, err := grpc.Dial(":"+portStr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
		if err != nil {
			log.Println("Could not connect: %s", err)
			continue
		}
		// Defer means: When this function returns, call this method (meaning, one main is done, close connection)
		defer conn.Close()

		//  Create new Client from generated gRPC code from proto
		mainClient = rep.NewReplicationClient(conn)
		log.Printf("Connected to main server at port: " + portStr)
		break
	}

	// This is our input function, that also keeps the client alive.
	takeInput()

}

func takeInput() {
	for {
		log.Println("Do want to \"Read\" or \"Add\" to the dictionary?")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		log.Print("(You wrote: " + input + ")")

		if strings.ToLower(input) == "read" {
			read()
			continue
		}
		if strings.ToLower(input) == "add" {
			add()
			continue
		}
		log.Println("Can't understand input, try again!")

	}
}

func read() {
	log.Println("Write a key corresponding to a value")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	log.Print("(You wrote: " + input + ")")

	mainClient.Read(context.Background(), input)

}

func add() {
	log.Println("Write the key you want to update")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	log.Println("Write the value you want to use")
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)

	log.Print("(You want to update: Key:" + key + ": Value:" + value + ")")
	mess, err := mainClient.Add(context.Background(), key, value)

	if err != nil {
		log.Fatalf("Something went wrong %s", err)
	}
	if mess.Success {
		log.Println("Succesful update!")
		return
	}
	log.Println("Unsuccesful update!")

}
