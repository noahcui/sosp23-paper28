package kvstore

import (
	logger "github.com/sirupsen/logrus"
	"github.com/sosp23/replicated-store/go/config"
	tcp "github.com/sosp23/replicated-store/go/multipaxos/network"
)

const (
	NotFound string = "key not found"
	Empty           = ""
)

type KVResult struct {
	Ok    bool
	Value string
}

type KVStore interface {
	Get(key string) *string
	Put(key string, value string) bool
	Del(key string) bool
	Close()
}

func CreateStore(config config.Config) KVStore {
	if config.Store == "mem" {
		return NewMemKVStore()
	} else {
		logger.Panic("no match kvstore")
		return nil
	}
}

func Execute(cmd *tcp.Command, store KVStore) KVResult {
	if cmd.Type == tcp.Get {
		value := store.Get(cmd.Key)
		if value != nil {
			return KVResult{Ok: true, Value: *value}
		} else {
			return KVResult{Ok: false, Value: NotFound}
		}
	}

	if cmd.Type == tcp.Put {
		if store.Put(cmd.Key, cmd.Value) {
			return KVResult{Ok: true, Value: Empty}
		}
		return KVResult{Ok: false, Value: NotFound}
	}

	if cmd.Type != tcp.Del {
		panic("Command type not Del")
	}

	if store.Del(cmd.Key) {
		return KVResult{Ok: true, Value: Empty}
	}
	return KVResult{Ok: false, Value: NotFound}
}
