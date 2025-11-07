package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Consumer ConsumerConfig `yaml:"consumer"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	HTTPPort int    `yaml:"http_port"`
	Name     string `yaml:"name"`
}

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	VHost        string `yaml:"vhost"`
	QueueName    string `yaml:"queue_name"`
	ExchangeName string `yaml:"exchange_name"`
	ExchangeType string `yaml:"exchange_type"`
	RoutingKey   string `yaml:"routing_key"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Workers       int  `yaml:"workers"`
	PrefetchCount int  `yaml:"prefetch_count"`
	AutoAck       bool `yaml:"auto_ack"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if config.Server.HTTPPort == 0 {
		config.Server.HTTPPort = 8080
	}
	if config.RabbitMQ.VHost == "" {
		config.RabbitMQ.VHost = "/"
	}
	if config.Consumer.Workers == 0 {
		config.Consumer.Workers = 3
	}
	if config.Consumer.PrefetchCount == 0 {
		config.Consumer.PrefetchCount = 10
	}

	return &config, nil
}

// GetRabbitMQURL 获取 RabbitMQ 连接 URL
func (r *RabbitMQConfig) GetRabbitMQURL() string {
	// 对 vhost 进行 URL 编码：将 / 替换为 %2F
	// 例如：/myvhost -> %2Fmyvhost
	encodedVHost := strings.ReplaceAll(r.VHost, "/", "%2F")
	
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		url.QueryEscape(r.Username),
		url.QueryEscape(r.Password),
		r.Host,
		r.Port,
		encodedVHost)
}
