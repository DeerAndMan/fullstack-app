package snowflake

import (
	"sync"
	"time"
)

const (
	epoch         = int64(1704067200000) // 2024-01-01 00:00:00 UTC
	nodeBits      = 10
	sequenceBits  = 12
	nodeMax       = -1 ^ (-1 << nodeBits)
	sequenceMask  = -1 ^ (-1 << sequenceBits)
	nodeShift     = sequenceBits
	timestampShift = nodeBits + sequenceBits
)

type Node struct {
	mu        sync.Mutex
	timestamp int64
	node      int64
	sequence  int64
}

var defaultNode *Node

func Init(node int64) {
	defaultNode = &Node{node: node & int64(nodeMax)}
}

func NextID() int64 {
	if defaultNode == nil {
		Init(1)
	}
	return defaultNode.Generate()
}

func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == n.timestamp {
		n.sequence = (n.sequence + 1) & int64(sequenceMask)
		if n.sequence == 0 {
			for now <= n.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.sequence = 0
	}
	n.timestamp = now

	return (now-epoch)<<timestampShift | n.node<<nodeShift | n.sequence
}
