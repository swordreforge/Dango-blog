package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Producer Sarama 生产者实现
type Producer struct {
	producer sarama.SyncProducer
	config   *KafkaConfig
}

// NewProducer 创建 Kafka 生产者
func NewProducer(config *KafkaConfig) (KafkaProducer, error) {
	if config == nil {
		config = NewKafkaConfig(nil, "")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("Kafka producer created successfully, brokers: %v", config.Brokers)
	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// SendMessage 发送消息到 Kafka
func (p *Producer) SendMessage(ctx context.Context, topic, key string, value interface{}) error {
	var valueBytes []byte
	var err error

	switch v := value.(type) {
	case []byte:
		valueBytes = v
	case string:
		valueBytes = []byte(v)
	default:
		valueBytes, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(valueBytes),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message sent successfully to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	log.Println("Kafka producer closed")
	return nil
}