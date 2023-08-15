package replicant

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"sync"

	logger "github.com/sirupsen/logrus"
	"github.com/sosp23/replicated-store/go/multipaxos"
	pb "github.com/sosp23/replicated-store/go/multipaxos/network"
)

func parse(request string) *pb.Command {
	substrings := strings.SplitN(strings.TrimRight(request, "\n"), " ", 4)
	if len(substrings) < 3 {
		return nil
	}
	commandID := substrings[0]
	commandType := substrings[1]
	key := substrings[2]

	command := &pb.Command{ID: commandID, Key: key}

	if commandType == "get" {
		command.Type = pb.Get
	} else if commandType == "del" {
		command.Type = pb.Del
	} else if commandType == "put" {
		if len(substrings) != 4 {
			return nil
		}
		command.Type = pb.Put
		command.Value = substrings[3]
	} else {
		return nil
	}
	return command
}

type Client struct {
	id           int64
	reader       *bufio.Reader
	writer       *bufio.Writer
	socket       net.Conn
	multipaxos   *multipaxos.Multipaxos
	manager      *ClientManager
	isFromClient bool
	writerLock   sync.Mutex
}

func NewClient(id int64, conn net.Conn, mp *multipaxos.Multipaxos,
	manger *ClientManager, isFromClient bool) *Client {
	client := &Client{
		id:           id,
		reader:       bufio.NewReader(conn),
		writer:       bufio.NewWriter(conn),
		socket:       conn,
		multipaxos:   mp,
		manager:      manger,
		isFromClient: isFromClient,
	}
	return client
}

func (c *Client) Start() {
	for {
		request, err := c.reader.ReadString('\n')
		if err != nil {
			logger.Error(err, c)
			// panic(err)
			break
		}
		c.handleRequest(request)

	}
	c.manager.Stop(c.id)
}

func (c *Client) Stop() {
	c.socket.Close()
}

func (c *Client) handleRequest(request string) {
	if c.isFromClient {
		// fmt.Println(request)
		go c.handleClientRequest(request)
	} else {
		go c.handlePeerRequest(request)
	}
}

func (c *Client) handleClientRequest(line string) {
	command := parse(line)

	if command != nil {
		// result := c.multipaxos.Replicate(command, c.id)
		// if result.Type == multipaxos.Ok {
		// 	return
		// }
		// if result.Type == multipaxos.Retry {
		// 	c.Write("retry")
		ballot := c.multipaxos.Ballot()
		if multipaxos.IsLeader(ballot, c.multipaxos.Id()) {
			c.Write("leader is me")
		} else {
			// if result.Type != multipaxos.SomeElseLeader {
			// 	panic("Result is not someone_else_leader")
			// }
			// c.Write("leader is ...")
			// If forward failed, just retry.
			rslt := c.multipaxos.ForwardToLeader(command, c.id)
			if rslt.Type == multipaxos.Ok {
				return
			}
			logger.Infof("forward not accepted")
			c.Write("retry")
		}
	} else {
		c.Write("bad command")
	}
}

func (c *Client) handlePeerRequest(line string) {
	var request pb.Message
	err := json.Unmarshal([]byte(line), &request)
	if err != nil {
		return
	}

	msg := []byte(request.Msg)
	go func() {
		switch pb.MessageType(request.Type) {
		case pb.PREPAREREQUEST:
			var prepareRequest pb.PrepareRequest
			json.Unmarshal(msg, &prepareRequest)
			prepareResponse := c.multipaxos.Prepare(prepareRequest)
			responseJson, _ := json.Marshal(prepareResponse)
			tcpMessage, _ := json.Marshal(pb.Message{
				Type:      uint8(pb.PREPARERESPONSE),
				ChannelId: request.ChannelId,
				Msg:       string(responseJson),
			})
			c.Write(string(tcpMessage))
		case pb.ACCEPTREQUEST:
			var acceptRequest pb.AcceptRequest
			json.Unmarshal(msg, &acceptRequest)
			acceptResponse := c.multipaxos.Accept(acceptRequest)
			responseJson, _ := json.Marshal(acceptResponse)
			tcpMessage, _ := json.Marshal(pb.Message{
				Type:      uint8(pb.ACCEPTRESPONSE),
				ChannelId: request.ChannelId,
				Msg:       string(responseJson),
			})
			c.Write(string(tcpMessage))
		case pb.COMMITREQUEST:
			var commitRequest pb.CommitRequest
			json.Unmarshal(msg, &commitRequest)
			commitResponse := c.multipaxos.Commit(commitRequest)
			responseJson, _ := json.Marshal(commitResponse)
			tcpMessage, _ := json.Marshal(pb.Message{
				Type:      uint8(pb.COMMITRESPONSE),
				ChannelId: request.ChannelId,
				Msg:       string(responseJson),
			})
			c.Write(string(tcpMessage))
		case pb.FORWARD:
			var forwardRequest pb.Forward
			json.Unmarshal(msg, &forwardRequest)
			// response := c.multipaxos.Replicate(forwardRequest.Command, forwardRequest.ClientId)
			ballot := c.multipaxos.Ballot()
			response := multipaxos.Result{
				Type:   multipaxos.Retry,
				Leader: multipaxos.ExtractLeaderId(ballot),
			}

			if multipaxos.IsLeader(ballot, c.multipaxos.Id()) {
				response.Type = multipaxos.Ok
				go c.multipaxos.Replicate(forwardRequest.Command, forwardRequest.ClientId)
			}

			forwardResponse := pb.ForwardResponse{
				Type: pb.Reject,
			}
			if response.Type == multipaxos.Ok {
				forwardResponse.Type = pb.Ok
			}
			responseJson, _ := json.Marshal(forwardResponse)
			tcpMessage, _ := json.Marshal(pb.Message{
				Type:      uint8(pb.FORWARDRESPONSE),
				ChannelId: request.ChannelId,
				Msg:       string(responseJson),
			})
			c.Write(string(tcpMessage))
		}
	}()
}

func (c *Client) Write(response string) {
	c.writerLock.Lock()
	defer c.writerLock.Unlock()
	_, err := c.writer.WriteString(response + "\n")
	if err == nil {
		c.writer.Flush()
	}
}
