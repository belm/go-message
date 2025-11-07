# 使用示例

## 1. 启动 RabbitMQ

确保 RabbitMQ 服务正在运行：

```bash
# macOS
brew services start rabbitmq

# Linux
sudo systemctl start rabbitmq-server

# Docker
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

## 2. 启动 HTTP 接收服务（生产者）

```bash
go run main.go -type=producer -config=config/config.yaml
```

服务将在 `http://localhost:8080` 启动。

## 3. 启动消息消费服务

在另一个终端窗口：

```bash
go run main.go -type=consumer -config=config/config.yaml
```

## 4. 发送测试消息

```bash
# 发送订单消息
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "type": "order",
    "order_id": "12345",
    "amount": 99.99,
    "customer": "张三"
  }'

# 发送通知消息
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "type": "notification",
    "title": "系统通知",
    "content": "这是一条测试消息"
  }'

# 健康检查
curl http://localhost:8080/health
```

## 5. 自定义消息处理器示例

创建 `handler/custom_handler.go`:

```go
package handler

import (
	"log"
	"go-message/model"
)

type CustomHandler struct {
	// 添加你的依赖，例如数据库连接、外部服务客户端等
}

func NewCustomHandler() *CustomHandler {
	return &CustomHandler{}
}

func (h *CustomHandler) Handle(message *model.Message) error {
	log.Printf("自定义处理器收到消息: %s", message.ID)
	
	// 根据消息类型进行不同处理
	switch message.Type {
	case "order":
		return h.processOrder(message)
	case "notification":
		return h.processNotification(message)
	default:
		log.Printf("未知消息类型: %s", message.Type)
		return nil
	}
}

func (h *CustomHandler) processOrder(message *model.Message) error {
	// 实现订单处理逻辑
	// 例如：保存到数据库、调用支付接口等
	log.Printf("处理订单: %v", message.Payload)
	return nil
}

func (h *CustomHandler) processNotification(message *model.Message) error {
	// 实现通知处理逻辑
	// 例如：发送邮件、推送通知等
	log.Printf("处理通知: %v", message.Payload)
	return nil
}
```

然后在 `main.go` 中使用：

```go
// 替换默认处理器
messageHandler := handler.NewCustomHandler()
consumer := queue.NewConsumer(conn, &cfg.RabbitMQ, &cfg.Consumer, messageHandler, cfg.Consumer.Workers)
```

