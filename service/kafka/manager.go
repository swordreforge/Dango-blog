package kafka

import (
	"fmt"
	"log"
	"sync"
)

// StagedConsumerManager 分阶段消费者管理器
type StagedConsumerManager struct {
	consumers   map[string]StagedConsumer
	mu          sync.RWMutex
	brokers     []string
	baseGroupID string
}

var (
	stagedConsumerManager *StagedConsumerManager
	managerOnce           sync.Once
)

// InitStagedConsumerManager 初始化分阶段消费者管理器
func InitStagedConsumerManager(brokers []string, baseGroupID string) {
	managerOnce.Do(func() {
		if len(brokers) == 0 {
			brokers = []string{"localhost:9092"}
		}
		if baseGroupID == "" {
			baseGroupID = "myblog-consumer-group"
		}

		stagedConsumerManager = &StagedConsumerManager{
			consumers:   make(map[string]StagedConsumer),
			brokers:     brokers,
			baseGroupID: baseGroupID,
		}
		log.Printf("Staged consumer manager initialized, base group: %s", baseGroupID)
	})
}

// GetStagedConsumer 获取指定阶段的消费者
func GetStagedConsumer(stage string) (StagedConsumer, error) {
	if stagedConsumerManager == nil {
		return nil, fmt.Errorf("staged consumer manager not initialized")
	}

	stagedConsumerManager.mu.RLock()
	consumer, exists := stagedConsumerManager.consumers[stage]
	stagedConsumerManager.mu.RUnlock()

	if exists {
		return consumer, nil
	}

	// 创建新的分阶段消费者
	stagedConsumerManager.mu.Lock()
	defer stagedConsumerManager.mu.Unlock()

	// 双重检查
	if consumer, exists := stagedConsumerManager.consumers[stage]; exists {
		return consumer, nil
	}

	newConsumer, err := NewStagedConsumer(stagedConsumerManager.brokers, stagedConsumerManager.baseGroupID, stage)
	if err != nil {
		return nil, err
	}

	stagedConsumerManager.consumers[stage] = newConsumer
	return newConsumer, nil
}

// CloseStagedConsumer 关闭指定阶段的消费者
func CloseStagedConsumer(stage string) error {
	if stagedConsumerManager == nil {
		return fmt.Errorf("staged consumer manager not initialized")
	}

	stagedConsumerManager.mu.Lock()
	defer stagedConsumerManager.mu.Unlock()

	consumer, exists := stagedConsumerManager.consumers[stage]
	if !exists {
		return fmt.Errorf("consumer for stage %s not found", stage)
	}

	if err := consumer.Close(); err != nil {
		return err
	}

	delete(stagedConsumerManager.consumers, stage)
	return nil
}

// CloseAllStagedConsumers 关闭所有分阶段消费者
func CloseAllStagedConsumers() {
	if stagedConsumerManager == nil {
		return
	}

	stagedConsumerManager.mu.Lock()
	defer stagedConsumerManager.mu.Unlock()

	for stage, consumer := range stagedConsumerManager.consumers {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing staged consumer for stage %s: %v", stage, err)
		}
	}

	stagedConsumerManager.consumers = make(map[string]StagedConsumer)
	log.Println("All staged consumers closed")
}

// ListStages 列出所有已创建的阶段
func ListStages() []string {
	if stagedConsumerManager == nil {
		return []string{}
	}

	stagedConsumerManager.mu.RLock()
	defer stagedConsumerManager.mu.RUnlock()

	stages := make([]string, 0, len(stagedConsumerManager.consumers))
	for stage := range stagedConsumerManager.consumers {
		stages = append(stages, stage)
	}
	return stages
}