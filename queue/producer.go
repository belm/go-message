package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"go-message/config"
	"go-message/model"
)

// Producer 消息生产者
type Producer struct {
	conn   *Connection
	config *config.RabbitMQConfig
}

// NewProducer 创建新的消息生产者
func NewProducer(conn *Connection, cfg *config.RabbitMQConfig) *Producer {
	return &Producer{
		conn:   conn,
		config: cfg,
	}
}

// Publish 发布消息到队列
func (p *Producer) Publish(message *model.Message) error {
	// 验证消息
	if err := message.Validate(); err != nil {
		return fmt.Errorf("消息验证失败: %w", err)
	}

	// 将消息转换为 JSON
	body, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发布消息
	err = p.conn.Channel().Publish(
		p.config.ExchangeName, // 交换机名称（如果为空则使用默认交换机）
		p.config.RoutingKey,   // 路由键（如果使用默认交换机，则路由键等于队列名）
		false,                 // 强制
		false,                 // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)

	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}

	log.Printf("消息已发布: ID=%s, Type=%s", message.ID, message.Type)
	return nil
}

// PublishBatch 批量发布消息
func (p *Producer) PublishBatch(messages []*model.Message) error {
	for _, msg := range messages {
		if err := p.Publish(msg); err != nil {
			return fmt.Errorf("批量发布消息失败: %w", err)
		}
	}
	return nil
}
