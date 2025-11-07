package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go-message/model"
	"go-message/queue"
)

// HTTPServer HTTP 接收服务
type HTTPServer struct {
	producer *queue.Producer
	port     int
}

// NewHTTPServer 创建新的 HTTP 服务器
func NewHTTPServer(producer *queue.Producer, port int) *HTTPServer {
	return &HTTPServer{
		producer: producer,
		port:     port,
	}
}

// Start 启动 HTTP 服务器
func (s *HTTPServer) Start() error {
	// 注册路由
	http.HandleFunc("/message", s.handleMessage)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("HTTP 服务器启动在端口 %d", s.port)
	
	return http.ListenAndServe(addr, nil)
}

// handleMessage 处理消息接收请求
func (s *HTTPServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var requestData map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	// 构建消息对象
	message := &model.Message{
		ID:        uuid.New().String(),
		Type:      getStringValue(requestData, "type", "default"),
		Payload:   requestData,
		Timestamp: time.Now(),
		Source:    r.RemoteAddr,
		Metadata:  extractMetadata(r),
	}

	// 发布消息到队列
	if err := s.producer.Publish(message); err != nil {
		log.Printf("发布消息失败: %v", err)
		http.Error(w, fmt.Sprintf("发布消息失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response := map[string]interface{}{
		"success": true,
		"message_id": message.ID,
		"message": "消息已接收并加入队列",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleHealth 健康检查端点
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getStringValue 从 map 中获取字符串值
func getStringValue(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

// extractMetadata 从 HTTP 请求中提取元数据
func extractMetadata(r *http.Request) map[string]string {
	metadata := make(map[string]string)
	
	// 提取常见的 HTTP 头信息
	if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
		metadata["user_agent"] = userAgent
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		metadata["content_type"] = contentType
	}
	
	return metadata
}

