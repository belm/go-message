package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/streadway/amqp"
	"go-message/config"
	"go-message/model"
)

// MessageHandler 消息处理接口
// 用户可以实现此接口来自定义消息处理逻辑
type MessageHandler interface {
	Handle(message *model.Message) error
}

// Consumer 消息消费者
type Consumer struct {
	conn         *Connection
	rabbitmqCfg  *config.RabbitMQConfig
	consumerCfg  *config.ConsumerConfig
	handler      MessageHandler
	workers      int
	wg           sync.WaitGroup
	done         chan bool
}

// NewConsumer 创建新的消息消费者
func NewConsumer(conn *Connection, rabbitmqCfg *config.RabbitMQConfig, consumerCfg *config.ConsumerConfig, handler MessageHandler, workers int) *Consumer {
	return &Consumer{
		conn:        conn,
		rabbitmqCfg:  rabbitmqCfg,
		consumerCfg:  consumerCfg,
		handler:     handler,
		workers:     workers,
		done:        make(chan bool),
	}
}

// Start 启动消费者
func (c *Consumer) Start() error {
	// 设置 QoS（服务质量）
	err := c.conn.Channel().Qos(
		c.consumerCfg.PrefetchCount, // 预取数量
		0,                           // 预取大小（0 表示不限制）
		false,                       // 全局设置
	)
	if err != nil {
		return fmt.Errorf("设置 QoS 失败: %w", err)
	}

	// 注册消费者
	msgs, err := c.conn.Channel().Consume(
		c.rabbitmqCfg.QueueName, // 队列名称
		"",                      // 消费者标签（空字符串表示自动生成）
		c.consumerCfg.AutoAck,   // 自动确认
		false,                   // 排他性
		false,                   // 无本地
		false,                   // 无等待
		nil,                     // 参数
	)
	if err != nil {
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	log.Printf("消费者已启动，工作协程数: %d", c.workers)

	// 启动多个工作协程处理消息
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(msgs, i)
	}

	return nil
}

// worker 工作协程
func (c *Consumer) worker(msgs <-chan amqp.Delivery, id int) {
	defer c.wg.Done()

	log.Printf("工作协程 %d 已启动", id)

	for {
		select {
		case <-c.done:
			log.Printf("工作协程 %d 已停止", id)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("工作协程 %d: 消息通道已关闭", id)
				return
			}

			// 处理消息
			if err := c.processMessage(msg, id); err != nil {
				log.Printf("工作协程 %d 处理消息失败: %v", id, err)
				// 如果处理失败且未自动确认，则拒绝消息（可以配置重试或死信队列）
				if !c.consumerCfg.AutoAck {
					msg.Nack(false, true) // 重新入队
				}
			} else {
				// 处理成功，手动确认（如果未自动确认）
				if !c.consumerCfg.AutoAck {
					msg.Ack(false)
				}
			}
		}
	}
}

// processMessage 处理单条消息
func (c *Consumer) processMessage(delivery amqp.Delivery, workerID int) error {
	var message model.Message

	// 解析消息
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	log.Printf("工作协程 %d 收到消息: ID=%s, Type=%s", workerID, message.ID, message.Type)

	// 调用用户定义的处理函数
	if err := c.handler.Handle(&message); err != nil {
		return fmt.Errorf("处理消息失败: %w", err)
	}

	log.Printf("工作协程 %d 处理消息成功: ID=%s", workerID, message.ID)
	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() {
	log.Println("正在停止消费者...")
	close(c.done)
	c.wg.Wait()
	log.Println("消费者已停止")
}

