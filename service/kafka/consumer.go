package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/IBM/sarama"
)

// Consumer Sarama 消费者实现
type Consumer struct {
	consumer sarama.ConsumerGroup
	config   *KafkaConfig
	handler  map[string]func(*sarama.ConsumerMessage)
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConsumer 创建 Kafka 消费者
func NewConsumer(config *KafkaConfig) (KafkaConsumer, error) {
	if config == nil {
		config = NewKafkaConfig(nil, "")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Kafka consumer created successfully, brokers: %v, group: %s", config.Brokers, config.GroupID)
	return &Consumer{
		consumer: consumer,
		config:   config,
		handler:  make(map[string]func(*sarama.ConsumerMessage)),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Subscribe 订阅主题
func (c *Consumer) Subscribe(topic string, handler func(*sarama.ConsumerMessage)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handler[topic] = handler

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, os.Interrupt, os.Kill)

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-sigterm:
				log.Println("Received termination signal, stopping consumer...")
				return
			default:
				if err := c.consumer.Consume(c.ctx, []string{topic}, &ConsumerGroupHandler{handler: handler}); err != nil {
					log.Printf("Error from consumer: %v", err)
				}
			}
		}
	}()

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.cancel()
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	log.Println("Kafka consumer closed")
	return nil
}

// GetGroupID 获取消费者组 ID
func (c *Consumer) GetGroupID() string {
	return c.config.GroupID
}