package nid

import (
	"testing"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"

	"github.com/stretchr/testify/assert"
)

func TestNodeNamed(t *testing.T) {
	named, err := NewConsulNamed("127.0.0.1:8500")
	assert.NoErrorf(t, err, "create failed")

	nodeID, err := named.GetNodeID(&NameHolder{
		LocalPath:  "test",
		LocalIP:    "127.0.0.1:8500",
		ServiceKey: "atlas/nodeIds",
	})

	assert.NoErrorf(t, err, "failed to get node id")
	assert.NotEqualf(t, 0, nodeID, "node id is zero")
}

func TestNewBoltNamed(t *testing.T) {
	named, err := NewBoltNamed("./node.bolt")
	assert.NoErrorf(t, err, "create failed")

	nodeID, err := named.GetNodeID(&NameHolder{
		LocalPath:  "test",
		LocalIP:    "127.0.0.1:8500",
		ServiceKey: "atlas/nodeIds",
	})

	assert.NoErrorf(t, err, "failed to get node id")
	assert.NotEqualf(t, 0, nodeID, "node id is zero")
}

func TestWatcher(t *testing.T) {
	kvStore, err := libkv.NewStore(
		store.CONSUL,
		[]string{"127.0.0.1:8500"},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)

	if err != nil {
		return
	}

	ch := make(chan struct{}, 1)
	pairChan, err := kvStore.WatchTree("dir", ch)
	if err != nil {
		return
	}

	timer := time.NewTimer(time.Second)

	for {
		select {
		case pairs, ok := <-pairChan:
			if !ok {
				return
			}

			for _, pair := range pairs {
				t.Logf("%v,%v", pair.Key, string(pair.Value))
			}
		case <-timer.C:
			ch <- struct{}{}
			return
		}
	}
}
