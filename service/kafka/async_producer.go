package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// ProducerMessage 队列中的消息封装
type ProducerMessage struct {
	Message   *sarama.ProducerMessage
	Callback  func(*sarama.ProducerMessage, error)
	Timestamp time.Time
}

// AsyncProducer 异步生产者实现
type AsyncProducer struct {
	producer sarama.AsyncProducer
	config   *KafkaConfig
	queue    chan *ProducerMessage
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	stats    *ProducerStats
}

// ProducerStats 生产者统计信息
type ProducerStats struct {
	mu            sync.RWMutex
	totalSent     int64
	totalErrors   int64
	queueSize     int
	queueCapacity int
}

// NewAsyncProducer 创建异步生产者
func NewAsyncProducer(config *KafkaConfig, queueSize int) (QueuedProducer, error) {
	if config == nil {
		config = NewKafkaConfig(nil, "")
	}
	if queueSize <= 0 {
		queueSize = 1000
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	saramaConfig.Producer.Flush.Frequency = 100 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 100

	producer, err := sarama.NewAsyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create async Kafka producer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	asyncProducer := &AsyncProducer{
		producer: producer,
		config:   config,
		queue:    make(chan *ProducerMessage, queueSize),
		ctx:      ctx,
		cancel:   cancel,
		stats: &ProducerStats{
			queueCapacity: queueSize,
		},
	}

	// 启动消息处理协程
	asyncProducer.start()

	log.Printf("Async Kafka producer created successfully, brokers: %v, queue size: %d", config.Brokers, queueSize)
	return asyncProducer, nil
}

// start 启动异步生产者的消息处理
func (ap *AsyncProducer) start() {
	ap.wg.Add(1)
	go func() {
		defer ap.wg.Done()
		ap.processMessages()
	}()

	// 处理成功消息
	ap.wg.Add(1)
	go func() {
		defer ap.wg.Done()
		for msg := range ap.producer.Successes() {
			ap.stats.mu.Lock()
			ap.stats.totalSent++
			ap.stats.mu.Unlock()
			log.Printf("Message sent successfully to topic %s, partition %d, offset %d", msg.Topic, msg.Partition, msg.Offset)
		}
	}()

	// 处理错误消息
	ap.wg.Add(1)
	go func() {
		defer ap.wg.Done()
		for err := range ap.producer.Errors() {
			ap.stats.mu.Lock()
			ap.stats.totalErrors++
			ap.stats.mu.Unlock()
			log.Printf("Failed to send message to topic %s: %v", err.Msg.Topic, err.Err)
		}
	}()
}

// processMessages 处理队列中的消息
func (ap *AsyncProducer) processMessages() {
	for {
		select {
		case <-ap.ctx.Done():
			return
		case msg := <-ap.queue:
			ap.stats.mu.Lock()
			ap.stats.queueSize = len(ap.queue)
			ap.stats.mu.Unlock()

			ap.producer.Input() <- msg.Message

			if msg.Callback != nil {
				go msg.Callback(msg.Message, nil)
			}
		}
	}
}

// SendAsync 异步发送消息
func (ap *AsyncProducer) SendAsync(ctx context.Context, topic, key string, value interface{}) error {
	return ap.SendAsyncWithCallback(ctx, topic, key, value, nil)
}

// SendAsyncWithCallback 异步发送消息并设置回调
func (ap *AsyncProducer) SendAsyncWithCallback(ctx context.Context, topic, key string, value interface{}, callback func(*sarama.ProducerMessage, error)) error {
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

	producerMsg := &ProducerMessage{
		Message:   msg,
		Callback:  callback,
		Timestamp: time.Now(),
	}

	select {
	case ap.queue <- producerMsg:
		ap.stats.mu.Lock()
		ap.stats.queueSize = len(ap.queue)
		ap.stats.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("queue timeout: unable to enqueue message after 5 seconds")
	}
}

// GetQueueSize 获取当前队列大小
func (ap *AsyncProducer) GetQueueSize() int {
	ap.stats.mu.RLock()
	defer ap.stats.mu.RUnlock()
	return ap.stats.queueSize
}

// GetQueueCapacity 获取队列容量
func (ap *AsyncProducer) GetQueueCapacity() int {
	ap.stats.mu.RLock()
	defer ap.stats.mu.RUnlock()
	return ap.stats.queueCapacity
}

// Flush 刷新队列，等待所有消息发送完成
func (ap *AsyncProducer) Flush() error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("flush timeout after 30 seconds")
		case <-ticker.C:
			if ap.GetQueueSize() == 0 {
				return nil
			}
		}
	}
}

// Close 关闭异步生产者
func (ap *AsyncProducer) Close() error {
	ap.cancel()

	// 等待队列中的消息处理完成
	if err := ap.Flush(); err != nil {
		log.Printf("Warning: flush error during close: %v", err)
	}

	// 关闭队列
	close(ap.queue)

	// 关闭生产者
	if err := ap.producer.Close(); err != nil {
		return fmt.Errorf("failed to close async producer: %w", err)
	}

	// 等待所有协程结束
	ap.wg.Wait()

	log.Println("Async Kafka producer closed")
	return nil
}

// GetStats 获取生产者统计信息
func (ap *AsyncProducer) GetStats() map[string]interface{} {
	ap.stats.mu.RLock()
	defer ap.stats.mu.RUnlock()

	return map[string]interface{}{
		"total_sent":     ap.stats.totalSent,
		"total_errors":   ap.stats.totalErrors,
		"queue_size":     ap.stats.queueSize,
		"queue_capacity": ap.stats.queueCapacity,
	}
}