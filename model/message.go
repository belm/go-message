package model

import (
	"encoding/json"
	"time"
)

// Message 通用消息结构
// 可以根据实际业务需求扩展字段
type Message struct {
	ID        string                 `json:"id"`        // 消息ID
	Type      string                 `json:"type"`      // 消息类型，用于区分不同业务场景
	Payload   map[string]interface{} `json:"payload"`   // 消息内容，灵活的数据结构
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Source    string                 `json:"source"`    // 消息来源
	Metadata  map[string]string      `json:"metadata"`  // 元数据，用于传递额外信息
}

// ToJSON 将消息转换为 JSON 字符串
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 字符串解析消息
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// Validate 验证消息是否有效
func (m *Message) Validate() error {
	if m.ID == "" {
		return ErrMessageIDRequired
	}
	if m.Type == "" {
		return ErrMessageTypeRequired
	}
	return nil
}

// 错误定义
var (
	ErrMessageIDRequired   = &MessageError{Code: "MESSAGE_ID_REQUIRED", Message: "消息ID不能为空"}
	ErrMessageTypeRequired = &MessageError{Code: "MESSAGE_TYPE_REQUIRED", Message: "消息类型不能为空"}
)

// MessageError 消息错误
type MessageError struct {
	Code    string
	Message string
}

func (e *MessageError) Error() string {
	return e.Message
}
