package nid

import (
	"testing"

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
