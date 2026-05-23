package snowflake

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	mu  sync.Mutex
	node *snowflake.Node
)

// Init configures the snowflake node (call once at startup, nodeID 0-1023).
func Init(nodeID int64) error {
	mu.Lock()
	defer mu.Unlock()
	if node != nil {
		return nil
	}
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("snowflake node: %w", err)
	}
	node = n
	return nil
}

// NextString returns a decimal string ID.
func NextString() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if node == nil {
		return "", fmt.Errorf("snowflake not initialized")
	}
	return node.Generate().String(), nil
}

// NextInt64 returns numeric snowflake ID.
func NextInt64() (int64, error) {
	mu.Lock()
	defer mu.Unlock()
	if node == nil {
		return 0, fmt.Errorf("snowflake not initialized")
	}
	return node.Generate().Int64(), nil
}
