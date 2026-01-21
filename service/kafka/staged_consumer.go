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

// StagedConsumerImpl 分阶段消费者实现
type StagedConsumerImpl struct {
	consumer sarama.ConsumerGroup
	config   *KafkaConfig
	stage    string
	handler  map[string]func(*sarama.ConsumerMessage)
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStagedConsumer 创建分阶段消费者
func NewStagedConsumer(brokers []string, groupID, stage string) (StagedConsumer, error) {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}
	if groupID == "" {
		groupID = "myblog-consumer-group"
	}
	if stage == "" {
		stage = "default"
	}

	// 为每个阶段创建独立的消费者组 ID
	stagedGroupID := fmt.Sprintf("%s-%s", groupID, stage)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, stagedGroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create staged Kafka consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Staged Kafka consumer created successfully, brokers: %v, group: %s, stage: %s", brokers, stagedGroupID, stage)
	return &StagedConsumerImpl{
		consumer: consumer,
		config:   &KafkaConfig{Brokers: brokers, GroupID: stagedGroupID},
		stage:    stage,
		handler:  make(map[string]func(*sarama.ConsumerMessage)),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Subscribe 分阶段消费者订阅主题
func (sc *StagedConsumerImpl) Subscribe(topic string, handler func(*sarama.ConsumerMessage)) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.handler[topic] = handler

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, os.Interrupt, os.Kill)

		for {
			select {
			case <-sc.ctx.Done():
				return
			case <-sigterm:
				log.Printf("Received termination signal for stage %s, stopping consumer...", sc.stage)
				return
			default:
				if err := sc.consumer.Consume(sc.ctx, []string{topic}, &ConsumerGroupHandler{handler: handler}); err != nil {
					log.Printf("Error from staged consumer (stage %s): %v", sc.stage, err)
				}
			}
		}
	}()

	log.Printf("Stage %s subscribed to topic: %s", sc.stage, topic)
	return nil
}

// Close 关闭分阶段消费者
func (sc *StagedConsumerImpl) Close() error {
	sc.cancel()
	if err := sc.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close staged consumer: %w", err)
	}
	log.Printf("Staged Kafka consumer closed, stage: %s", sc.stage)
	return nil
}

// GetGroupID 获取分阶段消费者组 ID
func (sc *StagedConsumerImpl) GetGroupID() string {
	return sc.config.GroupID
}

// GetStage 获取阶段名称
func (sc *StagedConsumerImpl) GetStage() string {
	return sc.stage
}