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

type msgBody struct {
	Type  string `json:"type,omitempty"`
	MsgID int    `json:"msg_id,omitempty"`
	Delta int    `json:"delta,omitempty"`
}

func main() {
	n := maelstrom.NewNode()

	kv := maelstrom.NewSeqKV(n)

	n.Handle("init", func(msg maelstrom.Message) error {
		// Unmarshal the built-in Maelstrom init body format
		var body maelstrom.InitMessageBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		_, err := kv.ReadInt(context.Background(), Counter)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				err = kv.CompareAndSwap(context.Background(), Counter, 0, 0, true)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}

		log.Printf("Successfully initialized node %s", body.NodeID)

		return nil
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var body msgBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		count, err := kv.ReadInt(context.Background(), Counter)
		if err != nil {
			return err
		}

		err = kv.CompareAndSwap(context.Background(), Counter, count, body.Delta+count, false)
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
