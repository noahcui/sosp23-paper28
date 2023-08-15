package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
	client "github.com/sosp23/replicated-store/go/benchmarker/clients"
)

type benchmarker struct {
	cfg     *client.Config
	clients []*client.Client
	data    []client.Data
}

func (b *benchmarker) storeData(dir string, filename string) {
	fullPath := filepath.Join(dir, filename)

	// Create the directory if it does not exist
	dirPath := filepath.Dir(fullPath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		errDir := os.MkdirAll(dirPath, 0755)
		if errDir != nil {
			logger.Fatal("Error creating directory", errDir)
		}
	}

	file, err := os.Create(fullPath)
	if err != nil {
		logger.Fatal("Cannot create file", err)
	}
	defer file.Close()

	sort.Slice(b.data, func(i, j int) bool {
		return b.data[i].StartTime.Before(b.data[j].StartTime)
	})

	writer := csv.NewWriter(file)
	defer writer.Flush()
	header := []string{"ClientId", "cmdID", "startTime", "endTime", "latency"}
	writer.Write(header)
	for _, value := range b.data {
		var data []string
		data = append(data, strconv.FormatInt(value.Id, 10))
		data = append(data, value.CmdID)
		data = append(data, value.StartTime.String())
		data = append(data, value.EndTime.String())
		data = append(data, value.Latency.String())
		writer.Write(data)
	}
}

func (b *benchmarker) SortData(totalTime time.Duration) {
	sort.Slice(b.data, func(i, j int) bool {
		return b.data[i].Latency < b.data[j].Latency
	})

	latencies := make([]time.Duration, len(b.data))
	var totalLatency time.Duration
	for i, item := range b.data {
		latencies[i] = item.Latency
		totalLatency += item.Latency
	}
	mean := totalLatency / time.Duration(len(b.data))
	median := latencies[len(latencies)/2]
	p90 := latencies[int(float64(len(latencies))*0.90)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]
	p999 := latencies[int(float64(len(latencies))*0.999)]
	p9999 := latencies[int(float64(len(latencies))*0.9999)]
	p99999 := latencies[int(float64(len(latencies))*0.99999)]

	totalCommands := float64(len(b.data))
	throughput := totalCommands / totalTime.Seconds()

	fmt.Printf("Mean Latency: %v\n", mean)
	fmt.Printf("Median Latency: %v\n", median)
	fmt.Printf("90th percentile latency: %v\n", p90)
	fmt.Printf("95th percentile latency: %v\n", p95)
	fmt.Printf("99th percentile latency: %v\n", p99)
	fmt.Printf("99.9th percentile latency: %v\n", p999)
	fmt.Printf("99.99th percentile latency: %v\n", p9999)
	fmt.Printf("99.999th percentile latency: %v\n", p99999)
	fmt.Printf("Total Time: %v\n", totalTime)
	fmt.Printf("Total Commands: %v\n", totalCommands)
	fmt.Printf("Throughput: %v\n", throughput)
}

func (b *benchmarker) Run() {
	t0 := time.Now()
	for _, c := range b.clients {
		go c.Run()
	}
	time.Sleep(time.Duration(b.cfg.DurationS) * time.Second)
	t1 := time.Now()
	for _, c := range b.clients {
		c.Stop()
	}
	totalTime := time.Since(t0)
	logger.Infoln("finished running clients, calculating time data and gathering from clients")

	wg := sync.WaitGroup{}
	for _, c := range b.clients {
		wg.Add(1)
		go func(wg *sync.WaitGroup, c *client.Client) {
			c.CalculateTimeData()
			wg.Done()
		}(&wg, c)
	}
	wg.Wait()
	for _, c := range b.clients {
		b.data = append(b.data, c.GetData()...)
	}
	t2 := time.Now()
	logger.Infof("After %v, finished gathering data from clients, now sorting the data\n", t2.Sub(t1))
	b.SortData(totalTime)
}

func main() {
	configPath := flag.String("c", "../c++/config.json", "config path")
	rawDataDir := flag.String("dir", "", "Where to store the raw data, default is do not store")
	fileName := flag.String("f", "data.csv", "the file name of the raw data")

	flag.Parse()
	cfg, err := client.LoadConfig(*configPath)
	if err != nil {
		logger.Panic(err)
	}
	b := &benchmarker{
		cfg:     &cfg,
		clients: make([]*client.Client, 0),
		data:    make([]client.Data, 0),
	}

	for i := 0; i < int(cfg.NumClients); i++ {
		j := i % len(cfg.Peers)
		b.clients = append(b.clients, client.NewClient(int64(i), &cfg, j))
	}

	b.Run()

	if rawDataDir != nil && *rawDataDir != "" {
		b.storeData(*rawDataDir, *fileName)
	}
}
