package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// KafkaMessage Kafka 消息结构
type KafkaMessage struct {
	Topic   string
	Key     string
	Content interface{}
}

// KafkaProducer Kafka 生产者接口
type KafkaProducer interface {
	SendMessage(ctx context.Context, topic, key string, value interface{}) error
	Close() error
}

// QueuedProducer 队列生产者接口
type QueuedProducer interface {
	SendAsync(ctx context.Context, topic, key string, value interface{}) error
	SendAsyncWithCallback(ctx context.Context, topic, key string, value interface{}, callback func(*sarama.ProducerMessage, error)) error
	GetQueueSize() int
	GetQueueCapacity() int
	Flush() error
	Close() error
}

// KafkaConsumer Kafka 消费者接口
type KafkaConsumer interface {
	Subscribe(topic string, handler func(*sarama.ConsumerMessage)) error
	Close() error
	GetGroupID() string
}

// StagedConsumer 分阶段消费者接口
type StagedConsumer interface {
	Subscribe(topic string, handler func(*sarama.ConsumerMessage)) error
	Close() error
	GetGroupID() string
	GetStage() string
}

// ConsumerGroupHandler 消费者组处理器
type ConsumerGroupHandler struct {
	handler func(*sarama.ConsumerMessage)
}

// Setup 在会话开始前调用
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 在会话结束后调用
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消息
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.handler(message)
		session.MarkMessage(message, "")
	}
	return nil
}

// NewKafkaConfig 创建 Kafka 配置
func NewKafkaConfig(brokers []string, groupID string) *KafkaConfig {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}
	if groupID == "" {
		groupID = "myblog-consumer-group"
	}
	return &KafkaConfig{
		Brokers: brokers,
		GroupID: groupID,
	}
}

// 全局 Kafka 生产者和消费者实例
var (
	kafkaProducer     KafkaProducer
	asyncProducer     QueuedProducer
	kafkaConsumer     KafkaConsumer
	kafkaOnce         sync.Once
	asyncProducerOnce sync.Once
)

// InitKafka 初始化 Kafka 服务
func InitKafka(brokers []string, groupID string) error {
	var initErr error

	kafkaOnce.Do(func() {
		config := NewKafkaConfig(brokers, groupID)

		// 创建生产者
		producer, err := NewProducer(config)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize Kafka producer: %w", err)
			return
		}
		kafkaProducer = producer

		// 创建消费者
		consumer, err := NewConsumer(config)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize Kafka consumer: %w", err)
			return
		}
		kafkaConsumer = consumer
	})

	return initErr
}

// InitAsyncProducer 初始化异步生产者
func InitAsyncProducer(brokers []string, queueSize int) error {
	var initErr error

	asyncProducerOnce.Do(func() {
		config := NewKafkaConfig(brokers, "")

		producer, err := NewAsyncProducer(config, queueSize)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize async Kafka producer: %w", err)
			return
		}
		asyncProducer = producer
	})

	return initErr
}

// GetKafkaProducer 获取 Kafka 生产者实例
func GetKafkaProducer() KafkaProducer {
	return kafkaProducer
}

// GetAsyncProducer 获取异步生产者实例
func GetAsyncProducer() QueuedProducer {
	return asyncProducer
}

// GetKafkaConsumer 获取 Kafka 消费者实例
func GetKafkaConsumer() KafkaConsumer {
	return kafkaConsumer
}

// CloseKafka 关闭 Kafka 服务
func CloseKafka() {
	if kafkaProducer != nil {
		if err := kafkaProducer.Close(); err != nil {
			fmt.Printf("Error closing Kafka producer: %v\n", err)
		}
		kafkaProducer = nil
	}

	if kafkaConsumer != nil {
		if err := kafkaConsumer.Close(); err != nil {
			fmt.Printf("Error closing Kafka consumer: %v\n", err)
		}
		kafkaConsumer = nil
	}
}

// CloseAsyncProducer 关闭异步生产者
func CloseAsyncProducer() {
	if asyncProducer != nil {
		if err := asyncProducer.Close(); err != nil {
			fmt.Printf("Error closing async Kafka producer: %v\n", err)
		}
		asyncProducer = nil
	}
}
