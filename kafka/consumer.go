package main

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	//if len(os.Args) != 2 {
	//	fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n",
	//		os.Args[0])
	//	os.Exit(1)
	//}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "PLAINTEXT://192.168.1.5:19092",
		"security.protocol": "PLAINTEXT",
		"sasl.mechanisms":   "PLAIN",
		"acks":              "all",
		"group.id":          "kafka-go-getting-started",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
		os.Exit(1)
	}

	topic := "purchases"
	err = c.SubscribeTopics([]string{topic}, nil)
	// Set up a channel for handling Ctrl-C, etc

	// Process messages

	for {

		ev, err := c.ReadMessage(100 * time.Millisecond)
		if err != nil {
			fmt.Printf("consuming messages failed :%s", err)
			// Errors are informational and automatically handled by the consumer
			continue
		}
		fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
			*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))

	}

	c.Close()

}
