package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-message/config"
	"go-message/handler"
	"go-message/queue"
)

const (
	// ServiceTypeProducer HTTP 接收服务（生产者）
	ServiceTypeProducer = "producer"
	// ServiceTypeConsumer 消息消费服务
	ServiceTypeConsumer = "consumer"
)

func main() {
	// 解析命令行参数
	serviceType := flag.String("type", "", fmt.Sprintf("服务类型: %s (HTTP接收服务) 或 %s (消息消费服务)", ServiceTypeProducer, ServiceTypeConsumer))
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 验证服务类型
	if *serviceType != ServiceTypeProducer && *serviceType != ServiceTypeConsumer {
		log.Fatalf("错误: 必须指定服务类型 -type=%s 或 -type=%s", ServiceTypeProducer, ServiceTypeConsumer)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.Printf("启动服务类型: %s", *serviceType)

	// 根据服务类型启动不同的服务
	switch *serviceType {
	case ServiceTypeProducer:
		startProducerService(cfg)
	case ServiceTypeConsumer:
		startConsumerService(cfg)
	default:
		log.Fatalf("未知的服务类型: %s", *serviceType)
	}
}

// startProducerService 启动 HTTP 接收服务（生产者）
func startProducerService(cfg *config.Config) {
	// 创建 RabbitMQ 连接
	conn, err := queue.NewConnection(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer conn.Close()

	// 创建消息生产者
	producer := queue.NewProducer(conn, &cfg.RabbitMQ)

	// 创建并启动 HTTP 服务器
	httpServer := handler.NewHTTPServer(producer, cfg.Server.HTTPPort)

	// 监听系统信号，优雅关闭
	setupGracefulShutdown(func() {
		log.Println("正在关闭 HTTP 服务器...")
		conn.Close()
	})

	// 启动 HTTP 服务器（阻塞）
	if err := httpServer.Start(); err != nil {
		log.Fatalf("HTTP 服务器启动失败: %v", err)
	}
}

// startConsumerService 启动消息消费服务
func startConsumerService(cfg *config.Config) {
	// 创建 RabbitMQ 连接
	conn, err := queue.NewConnection(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer conn.Close()

	// 创建消息处理器（用户可以替换为自己的实现）
	messageHandler := handler.NewDefaultMessageHandler()

	// 创建并启动消费者
	consumer := queue.NewConsumer(conn, &cfg.RabbitMQ, &cfg.Consumer, messageHandler, cfg.Consumer.Workers)

	if err := consumer.Start(); err != nil {
		log.Fatalf("启动消费者失败: %v", err)
	}

	log.Println("消息消费服务已启动，等待消息...")

	// 监听系统信号，优雅关闭
	setupGracefulShutdown(func() {
		log.Println("正在停止消费者...")
		consumer.Stop()
		conn.Close()
	})

	// 保持程序运行
	select {}
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown(cleanup func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("收到信号: %v", sig)
		cleanup()
		os.Exit(0)
	}()
}

