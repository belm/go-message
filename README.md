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
- ✅ Docker 容器化部署支持
- ✅ Docker Compose 一键启动所有服务

## 项目结构

```
go-message/
├── main.go                 # 主入口文件
├── Dockerfile             # Docker 镜像构建文件
├── docker-compose.yml     # Docker Compose 编排文件
├── .dockerignore          # Docker 构建忽略文件
├── Makefile               # Make 构建脚本（支持多平台编译）
├── config/
│   ├── config.yaml        # 本地开发配置文件
│   ├── config.docker.yaml # Docker 环境配置文件
│   └── config.go          # 配置加载和管理
├── docker/
│   └── rabbitmq-definitions.json # RabbitMQ 初始化配置
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

## Docker 部署

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+（推荐使用 Docker Compose）

### 方式一：使用 Docker Compose（推荐）

Docker Compose 会自动启动所有服务（RabbitMQ、生产者、消费者），是最简单的部署方式。

#### 1. 启动所有服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f                    # 查看所有服务日志
docker-compose logs -f producer           # 查看生产者日志
docker-compose logs -f consumer          # 查看消费者日志
docker-compose logs -f rabbitmq           # 查看 RabbitMQ 日志
```

#### 2. 停止服务

```bash
# 停止所有服务（保留数据）
docker-compose stop

# 停止并删除容器（保留数据卷）
docker-compose down

# 停止并删除容器和数据卷（清理所有数据）
docker-compose down -v
```

#### 3. 重新构建镜像

```bash
# 重新构建镜像（代码更新后）
docker-compose build

# 重新构建并启动
docker-compose up -d --build
```

#### 4. 服务访问

- **HTTP 生产者服务**: http://localhost:8080
- **RabbitMQ 管理界面**: http://localhost:15672
  - 用户名: `admin`
  - 密码: `admin123`

#### 5. 测试服务

```bash
# 发送测试消息
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "message": "Hello Docker!"
  }'

# 健康检查
curl http://localhost:8080/health
```

### 方式二：单独使用 Dockerfile

如果需要单独构建和运行容器：

#### 1. 构建镜像

```bash
docker build -t go-message:latest .
```

#### 2. 启动生产者服务

```bash
docker run -d --name go-message-producer \
  -p 8080:8080 \
  -v $(pwd)/config/config.docker.yaml:/app/config/config.yaml:ro \
  --network go-message-network \
  go-message:latest ./go-message -type=producer -config=/app/config/config.yaml
```

#### 3. 启动消费者服务

```bash
docker run -d --name go-message-consumer \
  -v $(pwd)/config/config.docker.yaml:/app/config/config.yaml:ro \
  --network go-message-network \
  go-message:latest ./go-message -type=consumer -config=/app/config/config.yaml
```

### Docker 配置文件说明

#### docker-compose.yml

包含三个服务：
- **rabbitmq**: RabbitMQ 消息队列服务
- **producer**: HTTP 消息接收服务（生产者）
- **consumer**: 消息消费服务

#### config/config.docker.yaml

Docker 环境专用配置文件，主要区别：
- `rabbitmq.host`: 使用 Docker 服务名 `rabbitmq`（而不是 IP 地址）
- 其他配置与 `config.yaml` 相同

#### docker/rabbitmq-definitions.json

RabbitMQ 初始化配置，自动创建：
- 虚拟主机 `/myvhost`
- 用户权限配置
- 队列和交换机
- 绑定关系

### Docker 部署常见问题

#### 1. 虚拟主机权限错误

如果遇到 `no access to this vhost` 或 `vhost not found` 错误：

```bash
# 清理旧数据并重新启动
docker-compose down -v
docker-compose up -d
```

#### 2. 服务无法连接 RabbitMQ

确保所有服务在同一 Docker 网络中：
```bash
# 检查网络
docker network ls
docker network inspect go-message_go-message-network
```

#### 3. 端口冲突

如果 8080 或 15672 端口被占用，修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "8081:8080"  # 将主机端口改为 8081
```

#### 4. 查看详细日志

```bash
# 查看特定服务的详细日志
docker-compose logs --tail=100 producer
docker-compose logs --tail=100 consumer

# 实时跟踪日志
docker-compose logs -f --tail=50
```

#### 5. 进入容器调试

```bash
# 进入生产者容器
docker exec -it go-message-producer sh

# 进入消费者容器
docker exec -it go-message-consumer sh

# 进入 RabbitMQ 容器
docker exec -it go-message-rabbitmq sh
```

### 生产环境建议

1. **数据持久化**: RabbitMQ 数据已配置为持久化存储（Docker volume）
2. **资源限制**: 建议在 `docker-compose.yml` 中添加资源限制：
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1'
         memory: 512M
   ```
3. **健康检查**: 服务已配置健康检查，可通过 `docker-compose ps` 查看状态
4. **日志管理**: 建议配置日志驱动，避免日志文件过大
5. **安全配置**: 生产环境请修改默认密码和用户名

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

