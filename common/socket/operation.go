package socket

import (
	"log"
	"reflect"
)

// SessionConnected 连接成功事件
type SessionConnected struct {
}

// SessionAccepted 接收其他服务器的连接
type SessionAccepted struct {
}

// ServiceIdentifyACK 连接成功后服务器节点回复验证信息
type ServiceIdentifyACK struct {
	ServiceId       string
	ServiceName     string
	ServerStartTime uint64 // 当前服务器启动时间
}

// PingReq 心跳包
type PingReq struct {
}

// PingAck 心跳包回复
type PingAck struct {
}

// SessionClosed 连接关闭事件
type SessionClosed struct {
}

func init() {
	RegisterSystemMsg(&SystemMsg{MsgId: 1, typ: reflect.TypeOf((*ServiceIdentifyACK)(nil)).Elem()})
	RegisterSystemMsg(&SystemMsg{MsgId: 2, typ: reflect.TypeOf((*PingReq)(nil)).Elem()})
	RegisterSystemMsg(&SystemMsg{MsgId: 3, typ: reflect.TypeOf((*PingAck)(nil)).Elem()})
	log.Println("operation init success")
}
