package client

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	NumClients  int64    `json:"num_clients"`
	Peers       []string `json:"peers"`
	ThinkTimems int64    `json:"thinktime_ms"`
	WriteRatio  float64  `json:"write_ratio"`
	KeyNum      int      `json:"key_num"`
	ValueSize   int      `json:"value_size"`
	DurationS   int64    `json:"duration_s"`
}

func DefaultConfig(n int) Config {
	peers := make([]string, n)
	for i := 0; i < n; i++ {
		peers[i] = "127.0.0.1:" + strconv.Itoa(10000+i*1000)
	}

	config := Config{
		NumClients:  1,
		Peers:       peers,
		ThinkTimems: 100,
		WriteRatio:  0.5,
		KeyNum:      1000,
		ValueSize:   100,
		DurationS:   10,
	}
	return config
}

func LoadConfig(configPath string) (Config, error) {
	var config Config
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
