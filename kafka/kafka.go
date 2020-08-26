package kafka

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

var requiredAcksMap = map[int]sarama.RequiredAcks{
	0:  sarama.WaitForAll,
	1:  sarama.WaitForLocal,
	-1: sarama.NoResponse,
}

var once sync.Once
var producers = sync.Map{}
var consumers = sync.Map{}

// Producer 生产者获取或初始化
func Producer(requireAcks int, addrs []string, timeout time.Duration) (sarama.SyncProducer, error) {
	p, ok := producers.Load(addrs[0])
	if ok {
		return p.(sarama.SyncProducer), nil
	}

	config := sarama.NewConfig()
	requiredAcks, ok := requiredAcksMap[requireAcks]
	if !ok {
		requiredAcks = sarama.WaitForAll
	}
	config.Producer.RequiredAcks = requiredAcks
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	init := make(chan bool)

	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			log.Fatal("kafka connection timed out")
		case <-init:
			break
		}
	}()
	producer, err := sarama.NewSyncProducer(addrs, config)
	init <- true
	if err != nil {
		return nil, err
	}
	producers.Store(addrs[0], producer)
	return producer, nil
}

// Consumer 消费者获取或初始化
func Consumer(addrs []string, timeout time.Duration) (sarama.Consumer, error) {
	c, ok := consumers.Load(addrs[0])
	if ok {
		return c.(sarama.Consumer), nil
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	kafkaInit := make(chan bool)

	go func() {
		defer cancel()
		select {
		case <-timeoutCtx.Done():
			log.Fatal("kafka connection timed out")
		case <-kafkaInit:
			break
		}
	}()

	consumer, err := sarama.NewConsumer(addrs, nil)
	kafkaInit <- true
	if err != nil {
		return nil, err
	}
	consumers.Store(addrs[0], consumer)
	return consumer, nil
}
