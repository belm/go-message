package handler

import (
	"log"

	"go-message/model"
)

// DefaultMessageHandler 默认消息处理器
// 用户可以根据业务需求实现 MessageHandler 接口来自定义处理逻辑
type DefaultMessageHandler struct {
	// 可以在这里添加业务相关的字段
}

// NewDefaultMessageHandler 创建默认消息处理器
func NewDefaultMessageHandler() *DefaultMessageHandler {
	return &DefaultMessageHandler{}
}

// Handle 处理消息
// 这是一个示例实现，用户应该根据实际业务需求修改此方法
func (h *DefaultMessageHandler) Handle(message *model.Message) error {
	log.Printf("处理消息: ID=%s, Type=%s", message.ID, message.Type)

	// 这里实现具体的业务逻辑
	// 例如：保存到数据库、调用外部 API、发送通知等

	// 示例：打印消息内容
	log.Printf("消息内容: %+v", message.Payload)

	// 示例：根据消息类型进行不同处理
	switch message.Type {
	case "order":
		return h.handleOrder(message)
	case "notification":
		return h.handleNotification(message)
	default:
		log.Printf("未知消息类型: %s", message.Type)
		return nil
	}
}

// handleOrder 处理订单消息（示例）
func (h *DefaultMessageHandler) handleOrder(message *model.Message) error {
	log.Printf("处理订单消息: %s", message.ID)
	// 实现订单处理逻辑
	return nil
}

// handleNotification 处理通知消息（示例）
func (h *DefaultMessageHandler) handleNotification(message *model.Message) error {
	log.Printf("处理通知消息: %s", message.ID)
	// 实现通知处理逻辑
	return nil
}

// 用户可以实现自己的消息处理器
// 例如：
//
// type CustomMessageHandler struct {
//     db *sql.DB
// }
//
// func (h *CustomMessageHandler) Handle(message *model.Message) error {
//     // 自定义处理逻辑
//     return nil
// }
