package replicant

import (
	"net"
	"strconv"
	"strings"

	logger "github.com/sirupsen/logrus"
	"github.com/sosp23/replicated-store/go/config"
	"github.com/sosp23/replicated-store/go/kvstore"
	consensusLog "github.com/sosp23/replicated-store/go/log"
	"github.com/sosp23/replicated-store/go/multipaxos"
)

type Replicant struct {
	id            int64
	log           *consensusLog.Log
	ipPort        string
	Multipaxos    *multipaxos.Multipaxos
	clientManager *ClientManager
	peerManager   *ClientManager
	peerListener  net.Listener
	acceptor      net.Listener
}

func NewReplicant(config config.Config) *Replicant {
	r := &Replicant{}
	r.id = config.Id
	r.ipPort = config.Peers[config.Id]
	r.log = consensusLog.NewLog(kvstore.CreateStore(config))
	r.peerListener, _ = net.Listen("tcp", r.ipPort)
	r.Multipaxos = multipaxos.NewMultipaxos(r.log, config)
	numPeers := int64(len(config.Peers))
	r.clientManager = NewClientManager(r.id, numPeers, r.Multipaxos, true, r)
	r.peerManager = NewClientManager(r.id, numPeers, r.Multipaxos, false, r)
	go r.StartPeerServer()
	return r
}

func (r *Replicant) executorTask() {
	for {
		id, result := r.log.Execute()
		if result == nil {
			break
		}
		client := r.clientManager.Get(id)
		if client != nil {
			client.Write(result.Value)
		}
	}
}

func (r *Replicant) serverTask() {
	for {
		conn, err := r.acceptor.Accept()
		if err != nil {
			logger.Error(err)
			break
		}
		r.clientManager.Start(conn)
	}
}

func (r *Replicant) peerServerTask() {
	logger.Infof("%v starting rpc server at %v", r.id, r.ipPort)
	for {
		client, err := r.peerListener.Accept()
		if err != nil {
			logger.Error(err)
			break
		}
		r.peerManager.Start(client)
	}
}

func (r *Replicant) Start() {
	r.Multipaxos.Start()
	r.StartExecutorTask()
	r.StartServerTask()
}

func (r *Replicant) Stop() {
	r.StopServer()
	r.StopExecutorThread()
	r.StopPeerServer()
	r.Multipaxos.Stop()
}

func (r *Replicant) StartServerTask() {
	pos := strings.Index(r.ipPort, ":")
	if pos == -1 {
		panic("no separator : in the acceptor port")
	}
	pos += 1
	port, err := strconv.Atoi(r.ipPort[pos:])
	if err != nil {
		panic("parsing acceptor port failed")
	}
	port += 1

	acceptor, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Infof("%v starting server at port %v\n", r.id, port)
	r.acceptor = acceptor
	r.serverTask()
}

func (r *Replicant) StopServer() {
	r.acceptor.Close()
	r.clientManager.StopAll()
}

func (r *Replicant) StartPeerServer() {
	go r.peerServerTask()
}

func (r *Replicant) StopPeerServer() {
	r.peerListener.Close()
	r.peerManager.StopAll()
}

func (r *Replicant) StartExecutorTask() {
	logger.Infof("%v starting executor thread\n", r.id)
	go r.executorTask()
}

func (r *Replicant) StopExecutorThread() {
	logger.Infof("%v stopping executor thread\n", r.id)
	r.log.Stop()
}

func (r *Replicant) LastExecuted() int64 {
	return r.log.LastExecuted()
}

func (r *Replicant) LastIndex() int64 {
	return r.log.LastIndex()
}

func (r *Replicant) LastCommitted() int64 {
	return r.log.LastCommitted()
}
