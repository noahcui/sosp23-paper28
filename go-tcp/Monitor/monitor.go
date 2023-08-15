package monitor

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sosp23/replicated-store/go/replicant"
)

type Monitor struct {
	r        *replicant.Replicant
	out_path string
	file     *os.File
	maxSize  int64 // in kilobytes
	fileSeq  int
	interval int // interval in milliseconds
}

// NewMonitor creates a new monitor
func NewMonitor(r *replicant.Replicant, path string, maxSize int64, interval int) *Monitor {
	m := &Monitor{
		r:        r,
		out_path: path,
		maxSize:  maxSize,
		fileSeq:  0,
		interval: interval,
	}
	return m
}

func (m *Monitor) Run() {
	ticker := time.NewTicker(time.Duration(m.interval) * time.Millisecond)
	for {
		<-ticker.C
		if err := m.ensureFileOpen(); err != nil {
			log.Printf("Failed to open file: %v", err)
			continue
		}
		writeValue := m.getInfo()
		if writeValue == "err" {
			continue
		}
		if _, err := m.file.WriteString(writeValue); err != nil {
			log.Printf("Failed to write to file: %v", err)
		}
	}
}

func (m *Monitor) getInfo() string {
	timenow := time.Now().Unix()
	challenSize := m.r.Multipaxos.SizeChannel()
	lastExecuted := m.r.LastExecuted()
	lastIndex := m.r.LastIndex()
	lastCommitted := m.r.LastCommitted()
	forwardings := m.r.Multipaxos.GetForwardings()
	forwardingLatencies := m.r.Multipaxos.GetForwardingLatencies()
	getmean := func(latencies []time.Duration) time.Duration {
		if len(latencies) == 0 {
			return 0
		}

		var sum time.Duration
		for _, latency := range latencies {
			sum += latency
		}
		return sum / time.Duration(len(latencies))
	}

	meanLatency := getmean(forwardingLatencies)

	toRet := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d\n", timenow, challenSize, lastExecuted, lastIndex, lastCommitted, forwardings, meanLatency)
	return toRet
}

func (m *Monitor) ensureFileOpen() error {
	if m.file != nil {
		info, err := m.file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}
		if m.maxSize <= 0 || info.Size() < m.maxSize*1024 {
			// File is either not too large, or we don't care about size
			return nil
		}
		// If the file is too large, we close it and set m.file to nil so a new file will be created
		m.file.Close()
		m.file = nil
		m.fileSeq++
	}

	var err error
	m.file, err = os.OpenFile(fmt.Sprintf("%s_%d_%d.csv", m.out_path, m.fileSeq, m.r.Multipaxos.Id()), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	writeValue := "time,channel_size,last_executed,last_index,last_committed,forwardings, forwarding_latencies\n"
	if _, err := m.file.WriteString(writeValue); err != nil {
		log.Printf("Failed to write to file: %v", err)
	}
	return err
}
