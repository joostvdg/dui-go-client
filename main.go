package main

import (
	"./feiwu"
	"./model"
	"./webserver"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	RED    = "\x1b[31;1m"
	GREEN  = "\x1b[32;1m"
	YELLOW = "\x1b[33;1m"
)

func listNetworkInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("We found an error")
		return
	}

	fmt.Printf("Found network interfaces:\n:")
	for _, networkInterface := range interfaces {
		fmt.Printf("%sName:%s, Index=%d, MAC:%d%s\n", YELLOW, networkInterface.Name, networkInterface.Index, networkInterface.HardwareAddr, YELLOW)
	}
}

func main() {
	fmt.Printf("Hello!\n")
	//listNetworkInterfaces()

	c := make(chan bool)
	var servers = map[string]model.FeiwuMessageOrigin{}
	webserverData := &webserver.WebserverData{Servers: servers}
	port := "7777"
	go webserver.StartServer(port, webserverData, c)

	fmt.Printf("> Started the web server on port %s, listening to DUI chatter\n", port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	listenToDui(webserverData)

	for i := 1; ; i++ { // this is still infinite
		t := time.NewTicker(time.Second * 30)
		select {
		case <-stop:
			fmt.Println("> Shutting down polling")
			break
		case <-t.C:
			listenToDui(webserverData)
			continue
		}
		break // only reached if the quitCh case happens
	}
	fmt.Println("> Shutting down webserver")
	c <- true
	if b := <-c; b {
		fmt.Println("> Webserver shut down")
	}
	fmt.Println("> Shut down app")
}

func listenToDui(webserverData *webserver.WebserverData) {
	addr, err := net.ResolveUDPAddr(feiwu.MEMBERSHIP_NETWORK, feiwu.MEMBERSHUP_GROUP_ADDRESS)
	if err != nil {
		fmt.Println("We found an error")
		return
	}
	conn, err3 := net.ListenMulticastUDP(feiwu.MEMBERSHIP_NETWORK, nil, addr)
	if err3 != nil {
		log.Fatal(err3)
	}
	readLoop(conn, webserverData)
}

func readLoop(c *net.UDPConn, webserverData *webserver.WebserverData) {

	log.Printf(" > [DUI Listener] readLoop: reading")

	buf := make([]byte, 10000)
	// TODO: optimize this!
	// currently it is 'eventually consistent' as in, it might not find all servers in one go
	for i := 0; i < 5; i++ {
		//n, cm, err1 := c.ReadFrom(buf)
		_, _, err := c.ReadFrom(buf)
		if err != nil {
			log.Printf(" > [DUI Listener] readLoop: ReadFrom: error %v\n", err)
			break
		}

		//var name string
		//ifi, err2 := net.InterfaceByName(cm.Network())
		//if err2 != nil {
		//	log.Printf("readLoop: unable to solve ifIndex=%s: error: %v\n", cm.Network(), err2)
		//}
		//
		//if ifi == nil {
		//	name = "ifname?"
		//} else {
		//	name = ifi.Name
		//}

		server := processFeiwuMessage(buf)
		webserverData.UpdateServers(server)
	}

	log.Printf(" > [DUI Listener] readLoop: exiting")
}
func processFeiwuMessage(rawMessage []byte) model.FeiwuMessageOrigin {
	bytesProcessed := 0
	feiwuHeader, bytesProcessed := getHeader(rawMessage, bytesProcessed, feiwu.HEADER_SIZE_FEIWU)
	messageTypeHeader, bytesProcessed := getHeader(rawMessage, bytesProcessed, feiwu.HEADER_SIZE_MESSAGE_TYPE)
	messageOriginSizeHeader, bytesProcessed := getHeader(rawMessage, bytesProcessed, feiwu.HEADER_SIZE_MESSAGE_ORIGIN_SIZE)
	messageSizeHeader, bytesProcessed := getHeader(rawMessage, bytesProcessed, feiwu.HEADER_SIZE_MESSAGE_SIZE)

	// 1 - PROTOCOL HEADER
	if feiwuHeader[0] == feiwu.FEIWU_HEADER_BYTE_MARKLER && feiwuHeader[1] == feiwu.FEIWU_HEADER_BYTE_MARKLER {
		log.Printf("> FEIWU message detected, processing...\n")
	} else {
		return model.FeiwuMessageOrigin{}
	}
	messageType := model.FeiwuMessageTypes[messageTypeHeader[1]]

	// 3 - MESSAGE ORIGIN SIZE
	messageOriginSize := binary.BigEndian.Uint32(messageOriginSizeHeader)
	//fmt.Printf(" - Message Origin Size: %v\n", messageOriginSize)

	// 4- MESSAGE PAYLOAD SIZE
	messageSize := binary.BigEndian.Uint32(messageSizeHeader)

	// 5 - MESSAGE ORIGIN DATA
	bytesToReadTo := bytesProcessed + int(messageOriginSize)
	rawMessageOriginData := rawMessage[bytesProcessed:bytesToReadTo]
	messageOriginRaw := string(rawMessageOriginData)
	bytesProcessed = bytesToReadTo

	messageOriginSplit := strings.Split(messageOriginRaw, ",")
	messageOrigin := model.FeiwuMessageOrigin{
		HostName:    messageOriginSplit[0],
		HostIP:      messageOriginSplit[1],
		ServerName:  messageOriginSplit[2],
		Role:        messageOriginSplit[3],
		LastSeen:    time.Now().String(),
		LastSeenRaw: time.Now(),
	}

	// 6 - MESSAGE DATA
	bytesToReadTo = bytesProcessed + int(messageSize)
	rawMessageData := rawMessage[bytesProcessed:bytesToReadTo]
	message := string(rawMessageData)
	bytesProcessed = bytesToReadTo

	// 7 - MESSAGE DIGEST
	messageDigest, bytesProcessed := getHeader(rawMessage, bytesProcessed, feiwu.HEADER_SIZE_MESSAGE_DIGEST)
	feiwuMessage := model.FeiwuMessage{
		Message:       message,
		MessageOrigin: messageOrigin,
		MessageDigest: messageDigest,
		MessageType:   messageType,
		Received:      time.Now().String(),
		ReceivedRaw:   time.Now(),
	}
	fmt.Printf(" - Message: %+v\n", feiwuMessage)
	return messageOrigin
}

func getHeader(rawMessage []byte, start int, length int) ([]byte, int) {
	end := start + length
	return rawMessage[start:end], end
}
