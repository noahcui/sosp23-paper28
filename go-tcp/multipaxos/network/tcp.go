package network

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/textproto"
	"sync"
	"time"
)

type Message struct {
	Type      uint8
	ChannelId uint64
	Msg       string
}

type ChannelMap struct {
	sync.Mutex
	Channels map[uint64]chan string
}

// func handleOutgoingRequests(stream net.Conn, requestChan chan string) {
// 	for request := range requestChan {
// 		request = request + "\n"
// 		requestBuffer := []byte(request)
// 		_, err := stream.Write(requestBuffer)
// 		if err != nil {
// 			return
// 		}
// 	}
// }

func handleOutgoingRequests(stream net.Conn, requestChan chan string) {
	const writeTimeout = 50 * time.Millisecond // TODO: go to config at sometime

	for request := range requestChan {
		request = request + "\n"
		requestBuffer := []byte(request)

		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()

		done := make(chan bool, 1)
		go func() {
			_, err := stream.Write(requestBuffer)
			if err != nil {
				log.Printf("Failed to write to stream: %v", err)
			}
			done <- true
		}()

		select {
		case <-ctx.Done():
			log.Printf("Request timed out: %v", ctx.Err())
		case <-done:
		}
	}
}

func handleIncomingResponses(stream net.Conn, channels *ChannelMap) {
	reader := textproto.NewReader(bufio.NewReader(stream))
	for {
		line, err := reader.ReadLineBytes()
		if err != nil {
			break
		}
		var response Message
		err = json.Unmarshal(line, &response)
		channels.Lock()
		if responseChan, ok := channels.Channels[response.ChannelId]; ok {
			responseChan <- response.Msg
		}
		channels.Unlock()
	}
	channels.Lock()
	defer channels.Unlock()
	for key := range channels.Channels {
		delete(channels.Channels, key)
	}
}

type TcpLink struct {
	requestChan chan string
}

func NewTcpLink(addr string, channels *ChannelMap) *TcpLink {
	var stream net.Conn
	for {
		var err error
		stream, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
	}
	go handleIncomingResponses(stream, channels)
	requestChan := make(chan string)
	go handleOutgoingRequests(stream, requestChan)
	return &TcpLink{
		requestChan: requestChan,
	}
}

func (t *TcpLink) Start() {

}

func (t *TcpLink) SendAwaitResponse(msgType MessageType, channelId uint64, msg string) {
	request, _ := json.Marshal(Message{
		Type:      uint8(msgType),
		ChannelId: channelId,
		Msg:       msg,
	})
	t.requestChan <- string(request)
}
