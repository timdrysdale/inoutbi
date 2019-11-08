package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	HOST    = "localhost"
	TYPE    = "tcp"
	BUFSIZE = 1024
)

var pIn, pOut, pBi string

func init() {
	flag.StringVar(&pIn, "in", "1234", "inward-only port number")
	flag.StringVar(&pOut, "out", "1235", "outward-only port number")
	flag.StringVar(&pBi, "bi", "1236", "bidirectional port number")
	flag.Parse()
}

func portSane(port int) bool {

	if port <= 1024 {
		return false
	}

	if port > 65535 {
		return false
	}

	return true
}

func portCheck(port string) {

	portNum, err := strconv.Atoi(port)
	if err != nil {
		// handle error
		fmt.Printf("Error specifying port %v\n", err)
		os.Exit(2)
	}
	if !portSane(portNum) {
		fmt.Printf("Invalid port %s because outside range 1025-65535\n", port)
		os.Exit(1)
	}
}

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	portCheck(pIn)
	portCheck(pOut)
	portCheck(pBi)

	inbi := make(chan []byte)
	biout := make(chan []byte)
	done := make(chan struct{})

	// Listen for connections to inward port
	cIn, err := net.Listen(TYPE, HOST+":"+pIn)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer cIn.Close()

	// Listen for connections to outward port
	cOut, err := net.Listen(TYPE, HOST+":"+pOut)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer cOut.Close()

	// Listen for incoming connections.
	cBi, err := net.Listen(TYPE, HOST+":"+pBi)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer cBi.Close()

	go handleInward(cIn, inbi)
	go handleOutward(cOut, biout)
	go handleBiward(cBi, inbi, biout)

	go func() {
		<-sigs
		close(done)
	}()

	<-done

	cIn.Close()
	cOut.Close()
	cBi.Close()

	close(inbi)
	close(biout)

}

func handleInward(l net.Listener, inbi chan []byte) {
	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine
		go handleInwardRequest(conn, inbi)
	}
}

func handleOutward(l net.Listener, biout chan []byte) {
	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine
		go handleOutwardRequest(conn, biout)
	}
}

func handleBiward(l net.Listener, inbi chan []byte, biout chan []byte) {
	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine
		go handleBiwardRequest(conn, inbi, biout)
	}
}

//write from port to channel
func handleInwardRequest(conn net.Conn, inbi chan []byte) {
	for {

		buf := make([]byte, BUFSIZE)

		_, err := conn.Read(buf)
		if err != nil {
			return
		}
		inbi <- buf
	}

}

//write from channel to port
func handleOutwardRequest(conn net.Conn, biout chan []byte) {
	for {

		buf := <-biout

		_, err := conn.Write(buf)
		if err != nil {
			return
		}
	}

}

//write from channel to port
func handleBiwardRequest(conn net.Conn, inbi chan []byte, biout chan []byte) {

	// send from bidirectional port to outward port
	go func() {

		buf := make([]byte, BUFSIZE)

		for {

			_, err := conn.Read(buf)
			if err != nil {
				return
			}

			biout <- buf
		}
	}()

	// send from inward port to bidirectional port
	go func() {

		for {

			buf := <-inbi

			_, err := conn.Write(buf)
			if err != nil {
				return
			}

		}
	}()

}
