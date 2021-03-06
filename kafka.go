package main

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaClient interface {
	Events() chan kafka.Event
	SubscribeTopics([]string, kafka.RebalanceCb) error
}

type Kafka struct {
	consumer kafkaClient
}

func NewKafka(bootstrapServers, topic string) (k Kafka, err error) {
	k.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               bootstrapServers,
		"auto.offset.reset":               "earliest",
		"group.id":                        "kafka-to-redisearch",
		"session.timeout.ms":              30000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": false,
		"enable.partition.eof":            true,
	})

	if err != nil {
		return
	}

	err = k.consumer.SubscribeTopics([]string{topic}, nil)

	return
}

func (k Kafka) ConsumerLoop(c chan MessageWithEnvelope) (err error) {
	for ev := range k.consumer.Events() {
		switch ev.(type) {
		case *kafka.Message:
			msg, err := ParseMessage(ev.(*kafka.Message).Value)
			if err != nil {
				log.Printf("Kafka Loop: invalid message: %+v", err)

				continue
			}

			c <- msg

		case kafka.Error:
			return ev.(kafka.Error)

		default:
			log.Printf("Kafka Loop: %+v, %T", ev, ev)
		}
	}
	return
}
