package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	protocol := "unix"
	sockAddr := os.Args[1]

	cleanup(sockAddr)

	fmt.Println("Server started...")
	listener, err := net.Listen(protocol, sockAddr)
	if err != nil {
		fmt.Println("Error starting socket server: " + err.Error())
	}
	defer listener.Close()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	// Quitting if we receive a SIGINT and do cleanup
	go func() {
		<-quit
		fmt.Println("Quitting..")
		close(quit)
		cleanup(sockAddr)
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error listening to client: " + err.Error())
			continue
		}
		go echo(conn)
	}
}

func echo(conn net.Conn) {
	defer conn.Close()
	log.Printf("Connected: %s\n", conn.RemoteAddr().Network())
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Print("Message Received:", string(message))

		type Params struct {
			Message string `json:"message" validate:"required"`
		}

		type RequestFormat struct {
			Id     int    `json:"id"`
			Method string `json:"method" validate:"required"`
			Params Params `json:"params" validate:"required"`
		}

		var requestFormat RequestFormat

		if err := json.Unmarshal([]byte(message), &requestFormat); err != nil {
			log.Printf("Error unmarshalling request: %s\n", err.Error())
			return
		}

		type ResponseFormat struct {
			Id     int `json:"id"`
			Result struct {
				Message string `json:"message"`
			} `json:"result"`
		}

		var responseFormat ResponseFormat

		responseFormat.Id = requestFormat.Id
		responseFormat.Result.Message = requestFormat.Params.Message

		response, err := json.Marshal(responseFormat)
		log.Print(string(response))
		if err != nil {
			log.Printf("Error marshalling response: %s\n", err.Error())
			return
		}
		conn.Write([]byte(string(response) + "\n"))
	}

}

func cleanup(sockAddr string) {
	if _, err := os.Stat(sockAddr); err == nil {
		if err := os.RemoveAll(sockAddr); err != nil {
			log.Fatal(err)
		}
	}
}
