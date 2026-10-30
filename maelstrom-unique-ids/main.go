package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	nodeKeyFilePath  = `%s.txt`
	keyFormat        = `%s-%d`
	keyBoundInterval = 5
)

type NodeKeyGenerator struct {
	nodeID        string
	keyUpperBound int
	currentKey    int
}

func NewNodeKeyGenerator(nodeID string) *NodeKeyGenerator {
	gen := &NodeKeyGenerator{nodeID: nodeID}
	gen.increaseUpperBound()
	return gen
}

func (g *NodeKeyGenerator) Generate() string {
	if g == nil || g.nodeID == "" {
		log.Fatal("node ID is not initialized")
	}
	log.Printf("node %s: generating key %d", g.nodeID, g.currentKey)
	g.currentKey++
	if g.currentKey == g.keyUpperBound {
		if err := g.increaseUpperBound(); err != nil {
			log.Fatal("failed to increase upper bound")
		}
	}
	return fmt.Sprintf(keyFormat, g.nodeID, g.currentKey)
}

func (g *NodeKeyGenerator) increaseUpperBound() error {
	nextUpperBound := keyBoundInterval
	currentKey := 0
	filePath := fmt.Sprintf(nodeKeyFilePath, g.nodeID)
	content, err := os.ReadFile(filePath)
	if err == nil {
		oldUpperBound, err := strconv.Atoi(string(content))
		if err != nil {
			log.Fatal("failed to initialize node key tracker")
		}
		if oldUpperBound < g.currentKey {
			log.Fatal("critical: key order broker")
		}
		currentKey = oldUpperBound
		nextUpperBound += oldUpperBound
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	if err = os.WriteFile(filePath, []byte(strconv.Itoa(nextUpperBound)), 0644); err != nil {
		log.Fatal("failed to create key file")
	}
	g.currentKey = currentKey
	g.keyUpperBound = nextUpperBound
	log.Printf("node %s: key upper bound increased to %d", g.nodeID, g.keyUpperBound)
	return nil
}

func main() {
	n := maelstrom.NewNode()

	var generator *NodeKeyGenerator

	n.Handle("init", func(msg maelstrom.Message) error {
		// Unmarshal the built-in Maelstrom init body format
		var body maelstrom.InitMessageBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		log.Printf("Successfully initialized node %s", body.NodeID)

		generator = NewNodeKeyGenerator(body.NodeID)

		// Let Maelstrom know initialization succeeded by sending the required "init_ok" reply
		return n.Reply(msg, map[string]string{"type": "init_ok"})
	})

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = generator.Generate()

		return n.Reply(msg, body)
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
