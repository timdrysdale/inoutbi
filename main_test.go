package main

import (
	"net"
	"testing"
	"time"
)

func TestInToBi(t *testing.T) {

	inExt, inInt := net.Pipe()
	outExt, outInt := net.Pipe()
	biExt, biInt := net.Pipe()

	inbi := make(chan []byte)
	biout := make(chan []byte)
	biwardRead := make(chan struct{})
	outwardRead := make(chan struct{})

	go handleInwardRequest(inInt, inbi)
	go handleOutwardRequest(outInt, biout)
	go handleBiwardRequest(biInt, inbi, biout)

	b := []byte("Test Message")

	inExt.Write(b)

	go func() {
		buf := make([]byte, BUFSIZE)
		outExt.Read(buf)
		close(outwardRead)
	}()

	go func() {
		buf := make([]byte, BUFSIZE)
		biExt.Read(buf)
		close(biwardRead)
	}()

	select {
	case <-biwardRead:
	case <-time.After(time.Millisecond):
		t.Error("Biward port timeout on message from inward port")
	}

	select {
	case <-outwardRead:
		t.Error("Outward port got a message from inward port")
	case <-time.After(time.Millisecond):
	}

	inExt.Close()
	inInt.Close()
	outExt.Close()
	outInt.Close()
	biExt.Close()
	biInt.Close()

}

func TestBiToOut(t *testing.T) {

	inExt, inInt := net.Pipe()
	outExt, outInt := net.Pipe()
	biExt, biInt := net.Pipe()

	inbi := make(chan []byte)
	biout := make(chan []byte)
	outwardRead := make(chan struct{})
	inwardRead := make(chan struct{})

	go handleInwardRequest(inInt, inbi)
	go handleOutwardRequest(outInt, biout)
	go handleBiwardRequest(biInt, inbi, biout)

	b := []byte("Test Message")

	biExt.Write(b)

	go func() {
		buf := make([]byte, BUFSIZE)
		outExt.Read(buf)
		close(outwardRead)
	}()

	go func() {
		buf := make([]byte, BUFSIZE)
		inExt.Read(buf)
		close(inwardRead)
	}()

	select {
	case <-outwardRead:
	case <-time.After(time.Millisecond):
		t.Error("Outward port timeout on message from inward port")
	}

	select {
	case <-inwardRead:
		t.Error("Inward port got a message from biward port")
	case <-time.After(time.Millisecond):
	}

	inExt.Close()
	inInt.Close()
	outExt.Close()
	outInt.Close()
	biExt.Close()
	biInt.Close()

}
