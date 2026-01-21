package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// PublishArticleEvent 发布文章事件
func PublishArticleEvent(ctx context.Context, eventType string, articleID int, title string) error {
	producer := GetKafkaProducer()
	if producer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"article_id": articleID,
		"title":      title,
		"timestamp":  ctx.Value("timestamp"),
	}

	return producer.SendMessage(ctx, "article-events", eventType, event)
}

// PublishCommentEvent 发布评论事件
func PublishCommentEvent(ctx context.Context, eventType string, commentID int, articleID int, content string) error {
	producer := GetKafkaProducer()
	if producer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type":  eventType,
		"comment_id":  commentID,
		"article_id":  articleID,
		"content":     content,
		"timestamp":   ctx.Value("timestamp"),
	}

	return producer.SendMessage(ctx, "comment-events", eventType, event)
}

// PublishUserEvent 发布用户事件
func PublishUserEvent(ctx context.Context, eventType string, userID int, username string) error {
	producer := GetKafkaProducer()
	if producer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"user_id":    userID,
		"username":   username,
		"timestamp":  ctx.Value("timestamp"),
	}

	return producer.SendMessage(ctx, "user-events", eventType, event)
}

// PublishArticleEventAsync 异步发布文章事件
func PublishArticleEventAsync(ctx context.Context, eventType string, articleID int, title string) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"article_id": articleID,
		"title":      title,
		"timestamp":  time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "article-events", eventType, event)
}

// PublishCommentEventAsync 异步发布评论事件
func PublishCommentEventAsync(ctx context.Context, eventType string, commentID int, articleID int, content string) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type":  eventType,
		"comment_id":  commentID,
		"article_id":  articleID,
		"content":     content,
		"timestamp":   time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "comment-events", eventType, event)
}

// PublishUserEventAsync 异步发布用户事件
func PublishUserEventAsync(ctx context.Context, eventType string, userID int, username string) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"user_id":    userID,
		"username":   username,
		"timestamp":  time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "user-events", eventType, event)
}

// PublishEventWithCallback 异步发布事件并设置回调
func PublishEventWithCallback(ctx context.Context, topic, eventType string, data interface{}, callback func(*sarama.ProducerMessage, error)) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	return producer.SendAsyncWithCallback(ctx, topic, eventType, data, callback)
}

// PublishAttachmentUploadEvent 发布附件上传事件
func PublishAttachmentUploadEvent(ctx context.Context, attachmentID int, fileName string, fileSize int64, fileType string, passageID int) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type":     "attachment.uploaded",
		"attachment_id":  attachmentID,
		"file_name":      fileName,
		"file_size":      fileSize,
		"file_type":      fileType,
		"passage_id":     passageID,
		"timestamp":      time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "attachment-events", "attachment.uploaded", event)
}

// PublishAttachmentDeleteEvent 发布附件删除事件
func PublishAttachmentDeleteEvent(ctx context.Context, attachmentID int, fileName string, passageID int) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type":     "attachment.deleted",
		"attachment_id":  attachmentID,
		"file_name":      fileName,
		"passage_id":     passageID,
		"timestamp":      time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "attachment-events", "attachment.deleted", event)
}

// PublishAttachmentUpdateEvent 发布附件更新事件
func PublishAttachmentUpdateEvent(ctx context.Context, attachmentID int, visibility string, showInPassage bool) error {
	producer := GetAsyncProducer()
	if producer == nil {
		return fmt.Errorf("Async Kafka producer not initialized")
	}

	event := map[string]interface{}{
		"event_type":      "attachment.updated",
		"attachment_id":   attachmentID,
		"visibility":      visibility,
		"show_in_passage": showInPassage,
		"timestamp":       time.Now().Unix(),
	}

	return producer.SendAsync(ctx, "attachment-events", "attachment.updated", event)
}