package main

import (
	"context"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	Counter = "counter"
)

func main() {
	n := maelstrom.NewNode()

	kv := maelstrom.NewSeqKV(n)

	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		count, err := kv.ReadInt(context.Background(), Counter)
		if err != nil {
			return err
		}

		delta, _ := body["delta"].(int)
		err = kv.CompareAndSwap(context.Background(), Counter, count, delta+count, true)
		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		count, err := kv.ReadInt(context.Background(), Counter)
		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": count,
		})
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
