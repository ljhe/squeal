package service

import (
	"common/iface"
	"log"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *SessionAccepted:
		// 服务器之间的心跳检测 (只能反应acceptor端的send和connector端的rcv是否正常)
		iv.Session().HeartBeat("server ping req")
		return nil
	case *SessionConnected:
		log.Println("服务器连接成功222", msg)
		return nil
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
