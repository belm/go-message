# Go 消息接收和处理系统

一个通用的、可扩展的消息接收和处理系统，使用 Go 语言实现。系统通过 HTTP 接口接收消息，将消息存入 RabbitMQ 队列进行削峰填谷，然后由消费者服务异步处理消息。

## 功能特性

- ✅ HTTP 服务接收 JSON 格式消息
- ✅ 消息存入 RabbitMQ 队列（削峰填谷）
- ✅ 独立的消费者服务处理消息
- ✅ 单入口文件，通过参数区分服务类型
- ✅ 配置文件管理，无需硬编码
- ✅ 通用设计，易于扩展和定制
- ✅ 支持多工作协程并发处理
- ✅ 优雅关闭机制

## 项目结构

```
go-message/
├── main.go                 # 主入口文件
├── config/
│   ├── config.yaml        # 配置文件
│   └── config.go          # 配置加载和管理
├── model/
│   └── message.go         # 消息模型定义
├── queue/
│   ├── rabbitmq.go        # RabbitMQ 连接管理
│   ├── producer.go        # 消息生产者
│   └── consumer.go        # 消息消费者
└── handler/
    ├── http_handler.go    # HTTP 请求处理
    └── message_handler.go # 消息处理逻辑
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置 RabbitMQ

确保已安装并运行 RabbitMQ。默认配置：
- 地址: localhost:5672
- 用户名: guest
- 密码: guest

### 3. 修改配置文件

编辑 `config/config.yaml`，根据实际情况修改配置：

```yaml
server:
  http_port: 8080

rabbitmq:
  host: "localhost"
  port: 5672
  username: "guest"
  password: "guest"
  queue_name: "message_queue"
```

### 4. 启动服务

**启动 HTTP 接收服务（生产者）：**
```bash
go run main.go -type=producer -config=config/config.yaml
```

**启动消息消费服务：**
```bash
go run main.go -type=consumer -config=config/config.yaml
```

## 使用说明

### 发送消息

向 HTTP 服务发送 POST 请求：

```bash
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "type": "order",
    "payload": {
      "order_id": "12345",
      "amount": 99.99
    }
  }'
```

### 消息格式

消息会自动添加以下字段：
- `id`: 自动生成的唯一消息ID
- `type`: 消息类型（从请求中获取，默认为 "default"）
- `payload`: 完整的请求数据
- `timestamp`: 消息时间戳
- `source`: 请求来源地址
- `metadata`: HTTP 请求元数据

### 自定义消息处理

系统提供了 `MessageHandler` 接口，你可以实现自己的消息处理逻辑：

1. 创建自定义处理器：

```go
package handler

import "go-message/model"

type CustomHandler struct {
    // 添加你的依赖字段
}

func (h *CustomHandler) Handle(message *model.Message) error {
    // 实现你的业务逻辑
    // 例如：保存到数据库、调用外部API等
    return nil
}
```

2. 在 `main.go` 中使用：

```go
// 替换默认处理器
messageHandler := &handler.CustomHandler{}
consumer := queue.NewConsumer(conn, &cfg.RabbitMQ, &cfg.Consumer, messageHandler, cfg.Consumer.Workers)
```

## 配置说明

### 服务配置

- `server.http_port`: HTTP 服务监听端口
- `server.name`: 服务名称

### RabbitMQ 配置

- `rabbitmq.host`: RabbitMQ 服务器地址
- `rabbitmq.port`: RabbitMQ 端口
- `rabbitmq.username`: 用户名
- `rabbitmq.password`: 密码
- `rabbitmq.vhost`: 虚拟主机
- `rabbitmq.queue_name`: 队列名称
- `rabbitmq.exchange_name`: 交换机名称（可选）
- `rabbitmq.exchange_type`: 交换机类型（direct, topic, fanout, headers）
- `rabbitmq.routing_key`: 路由键

### 消费者配置

- `consumer.workers`: 并发工作协程数量
- `consumer.prefetch_count`: 消息预取数量
- `consumer.auto_ack`: 是否自动确认消息（false 表示手动确认，更安全）

## API 端点

### POST /message

接收消息并加入队列。

**请求体：**
```json
{
  "type": "order",
  "order_id": "12345",
  "amount": 99.99
}
```

**响应：**
```json
{
  "success": true,
  "message_id": "uuid-string",
  "message": "消息已接收并加入队列"
}
```

### GET /health

健康检查端点。

**响应：**
```json
{
  "status": "ok",
  "timestamp": 1234567890
}
```

## 扩展指南

系统设计为通用架构，可以通过以下方式扩展：

1. **自定义消息处理逻辑**：实现 `MessageHandler` 接口
2. **添加新的消息类型**：在 `message_handler.go` 中添加处理分支
3. **添加中间件**：在 HTTP 处理器中添加认证、日志等中间件
4. **添加监控**：集成 Prometheus、Grafana 等监控工具
5. **添加重试机制**：在消费者中实现消息重试逻辑
6. **添加死信队列**：配置 RabbitMQ 死信队列处理失败消息

## 注意事项

- 确保 RabbitMQ 服务正常运行
- 生产环境建议设置 `auto_ack: false` 以确保消息可靠处理
- 根据实际负载调整 `workers` 和 `prefetch_count` 参数
- 建议使用环境变量或配置中心管理敏感配置信息

## 依赖包

- `github.com/streadway/amqp`: RabbitMQ 客户端
- `gopkg.in/yaml.v3`: YAML 配置文件解析
- `github.com/google/uuid`: UUID 生成

## 许可证

MIT License

