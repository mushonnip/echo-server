package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/go-playground/validator"
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

	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, conn)
	if err != nil {
		log.Println(err)
		return
	}

	incomingString := buf.String()
	fmt.Println("Request: " + incomingString)

	buf.Reset()

	type RequestFormat struct {
		Id     int    `json:"id"`
		Method string `json:"method" validate:"required"`
		Params struct {
			Message string `json:"message" validate:"required"`
		} `json:"params" validate:"required"`
	}

	// Unmarshal the incoming JSON
	var requestFormat RequestFormat
	err = json.Unmarshal([]byte(incomingString), &requestFormat)
	if err != nil {
		fmt.Println(err.Error())
	}

	validate := validator.New()
	err = validate.Struct(requestFormat)
	if err != nil {
		fmt.Println(err.Error())
		conn.Close() // disconnect the client
	}

	// Echo to the client
	// buf.WriteString(fmt.Sprintf(`{"id": %d, "result": {"message": "%s"}}\n`, requestFormat.Id, requestFormat.Params.Message))
	buf.Write([]byte(fmt.Sprintf(`{"id": %d, "result": {"message": "%s"}}`, requestFormat.Id, requestFormat.Params.Message)))

	_, err = io.Copy(conn, buf)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Response: ", fmt.Sprintf(`{"id":%d,"result":{"message":"%s"}}`, requestFormat.Id, requestFormat.Params.Message))
}

func cleanup(sockAddr string) {
	if _, err := os.Stat(sockAddr); err == nil {
		if err := os.RemoveAll(sockAddr); err != nil {
			log.Fatal(err)
		}
	}
}
