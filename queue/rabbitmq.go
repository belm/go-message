package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"go-message/config"
)

// Connection RabbitMQ 连接封装
type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.RabbitMQConfig
}

// NewConnection 创建新的 RabbitMQ 连接
func NewConnection(cfg *config.RabbitMQConfig) (*Connection, error) {
	url := cfg.GetRabbitMQURL()
	
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建 RabbitMQ 通道失败: %w", err)
	}

	// 声明交换机（如果配置了）
	if cfg.ExchangeName != "" {
		err = channel.ExchangeDeclare(
			cfg.ExchangeName, // 交换机名称
			cfg.ExchangeType, // 交换机类型
			true,             // 持久化
			false,            // 自动删除
			false,            // 内部使用
			false,            // 无等待
			nil,              // 参数
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("声明交换机失败: %w", err)
		}
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		cfg.QueueName, // 队列名称
		true,          // 持久化
		false,         // 自动删除
		false,         // 排他性
		false,         // 无等待
		nil,           // 参数
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("声明队列失败: %w", err)
	}

	// 绑定队列到交换机（如果配置了交换机）
	if cfg.ExchangeName != "" {
		err = channel.QueueBind(
			cfg.QueueName,   // 队列名称
			cfg.RoutingKey,  // 路由键
			cfg.ExchangeName, // 交换机名称
			false,            // 无等待
			nil,              // 参数
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("绑定队列到交换机失败: %w", err)
		}
	}

	log.Printf("成功连接到 RabbitMQ: %s", url)

	return &Connection{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

// Channel 获取通道
func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("关闭通道失败: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("关闭连接失败: %w", err)
		}
	}
	return nil
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	return c.conn.IsClosed()
}

