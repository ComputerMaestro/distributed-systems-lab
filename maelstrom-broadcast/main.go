package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type MessagesStore struct {
	messages []any
	m        sync.Mutex
	topology map[string][]string
}

func NewMessageStore() *MessagesStore {
	return &MessagesStore{}
}

func (ms *MessagesStore) AddMessage(message any) {
	ms.m.Lock()
	defer ms.m.Unlock()
	ms.messages = append(ms.messages, message)
}

func (ms *MessagesStore) GetMessages() []any {
	return ms.messages
}

func main() {
	store := NewMessageStore()
	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		store.AddMessage(body["message"])

		for _, nodeId := range n.NodeIDs() {
			if n.ID() != nodeId {
				n.Send(nodeId, map[string]any{
					"type":    "sync",
					"message": body["message"],
				})
			}
		}

		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		store.AddMessage(body["message"])

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": store.GetMessages(),
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
