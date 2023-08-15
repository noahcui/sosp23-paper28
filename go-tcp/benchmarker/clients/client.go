package client

import (
	"bufio"
	crand "crypto/rand"
	"encoding/csv"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/sosp23/replicated-store/go/multipaxos/network"
)

const (
	LEADERISME = "leader is me\n"
	RETRY      = "retry"
)

type Data struct {
	Id        int64
	CmdID     string
	StartTime time.Time
	EndTime   time.Time
	Latency   time.Duration
}

type Client struct {
	cfg         *Config
	peerID      int
	id          int64
	thinkTimems int64
	stopped     int32
	reader      *bufio.Reader
	writer      *bufio.Writer
	socket      net.Conn
	writeRatio  float64
	keyNum      int
	valueSize   int // in bytes
	commandID   int64
	startTimes  map[string]time.Time
	endTimes    map[string]time.Time
	mu          sync.Mutex
	data        []Data
}

type Command struct {
	Type  network.CommandType
	Key   string
	Value string
}

func NewClient(id int64, cfg *Config, peerID int) *Client {
	c := &Client{
		cfg:         cfg,
		peerID:      peerID,
		id:          id,
		thinkTimems: cfg.ThinkTimems,
		stopped:     0,
		writeRatio:  cfg.WriteRatio,
		keyNum:      1000,
		valueSize:   100,
		commandID:   0,
		startTimes:  make(map[string]time.Time),
		endTimes:    make(map[string]time.Time),
		data:        make([]Data, 0),
	}
	ip := strings.SplitN(cfg.Peers[peerID], ":", 2)[0]
	port := strings.SplitN(cfg.Peers[peerID], ":", 2)[1]
	c.Init(ip, port)
	return c
}

func (c *Client) Stop() {
	c.socket.Close()
	atomic.StoreInt32(&c.stopped, 1)
	// c.StoreTimeToCSVFile()
}

func (c *Client) GetStartTimes() map[string]time.Time {
	return c.startTimes
}

func (c *Client) GetEndTimes() map[string]time.Time {
	return c.endTimes
}

func (c *Client) GetData() []Data {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func (c *Client) CalculateTimeData() {
	c.mu.Lock()
	for id, end := range c.endTimes {
		// if id != "Test" && c.startTimes[id] != (time.Time{}) {
		if id != "Test" {
			start := c.startTimes[id]
			l := end.Sub(start)
			c.data = append(c.data, Data{c.id, id, start, end, l})
		}
		delete(c.startTimes, id)
		delete(c.endTimes, id)
	}
	c.mu.Unlock()
}

func (c *Client) StoreTimeToCSVFile() {
	file, err := os.Create("client" + strconv.Itoa(int(c.id)) + ".csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	c.mu.Lock()
	for id, start := range c.startTimes {
		end := c.endTimes[id]
		writer.Write([]string{id, strconv.FormatInt(start.UnixNano(), 10), strconv.FormatInt(end.UnixNano(), 10)})
	}
	c.mu.Unlock()
}

func (c *Client) Init(ip string, port string) {
	socket, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		panic(err)
	}
	c.socket = socket
	c.reader = bufio.NewReader(socket)
	c.writer = bufio.NewWriter(socket)
	atomic.StoreInt32(&c.stopped, 0)
	c.WarmUp()
}

func (c *Client) WarmUp() {
	for i := 0; i < 10; i++ {
		c.SendWarmUpRequest()
	}
}

func (c *Client) getMsg(cmdTypeStr string) (string, string) {
	id, key, value := c.getRandKV()
	return id, cmdTypeStr + " " + key + " " + value + "\n"
}

func (c *Client) SendWarmUpRequest() {
	msg := ""
	id := ""
	if c.writeRatio > 0 && c.writeRatio > rand.Float64() {
		id, msg = c.getMsg("get")
	} else {
		id, msg = c.getMsg("put")
	}
	id = "Test"
	start := time.Now()
	c.writer.WriteString(id + " " + msg)
	c.writer.Flush()
	c.mu.Lock()
	c.startTimes[id] = start
	c.mu.Unlock()

}

func (c *Client) SendRequest() {
	msg := ""
	id := ""
	if c.writeRatio > 0 && c.writeRatio > rand.Float64() {
		id, msg = c.getMsg("get")
	} else {
		id, msg = c.getMsg("put")
	}
	start := time.Now()
	c.writer.WriteString(id + " " + msg)
	c.writer.Flush()
	c.mu.Lock()
	c.startTimes[id] = start
	c.mu.Unlock()

}

func (c *Client) readerThread() {
	for atomic.LoadInt32(&c.stopped) == 0 {
		result, err := c.reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if result == LEADERISME {
			logger.Infoln("talking to leader, reiniting")
			c.Stop()
			c.peerID = rand.Intn(3)
			c.Init(strings.SplitN(c.cfg.Peers[c.peerID], ":", 2)[0], strings.SplitN(c.cfg.Peers[c.peerID], ":", 2)[1])
		} else if result == RETRY {
			// we are not retrying here. Just ignore it
		} else {
			substrings := strings.SplitN(strings.TrimRight(result, "\n"), " ", 2)
			end := time.Now()
			c.mu.Lock()
			// if there a duplicated msg, do not update the end time
			if _, ok := c.endTimes[substrings[0]]; !ok {
				c.endTimes[substrings[0]] = end
				// fmt.Println("not found", substrings[0])
			}
			c.mu.Unlock()
		}

	}
}

func (c *Client) Run() {
	go c.readerThread()
	ticker := time.NewTicker(time.Duration(c.thinkTimems) * time.Millisecond)
	for atomic.LoadInt32(&c.stopped) == 0 {
		<-ticker.C
		go c.SendRequest()
	}
}

func (c *Client) getRandKV() (string, string, string) {
	id := atomic.AddInt64(&c.commandID, 1)
	key := fmt.Sprint(rand.Intn(c.keyNum))
	// value := fmt.Sprint(rand.Intn(c.valueSize))
	value, _ := RandomString(c.valueSize)

	return fmt.Sprint(id), key, value
}

func RandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := range ret {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
