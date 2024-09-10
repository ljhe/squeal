package service

import (
	"common"
	"common/baseserver"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/socket"
	"common/util"
	"log"
	"reflect"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *socket.SessionAccepted:
		log.Printf("receive SessionAccepted success. session:%d \n", iv.Session().GetId())
		return nil
	case *socket.SessionConnected:
		// 从内存中的etcd获取服务器信息
		ctx := iv.Session().Node().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		if ctx.RawContextData(common.ContextSetEtcdKey, &ed) {
			prop := iv.Session().Node().(common.ServerNodeProperty)
			// 连接上服务器节点后 发送确认信息 告诉对端自己的服务器信息
			iv.Session().Send(&socket.ServiceIdentifyACK{
				ServiceId:       util.GenServiceId(prop),
				ServiceName:     prop.GetName(),
				ServerStartTime: util.GetCurrentTimeMs(),
			})
			// 添加远程的服务器节点信息到本地
			baseserver.AddServiceNode(iv.Session(), ed.Id, ed.Name, "local")
			log.Printf("send ServiceIdentifyACK [%v]->[%v] sessionId=%v \n",
				util.GenServiceId(prop), ed.Id, iv.Session().GetId())
		} else {
			log.Println("connector connect err. etcd not exist", msg)
		}
		return nil
	case *socket.ServiceIdentifyACK:
		// 来自其他服务器的连接确认信息
		log.Printf("receive ServiceIdentifyACK from [%v]  sessionId:%v \n", msg.ServiceId, iv.Session().GetId())
		// 重连时会有问题 重连上来时 但是上一个连接还未移除(正在移除中) 导致重连失败(想连接的没连接上 该移除的正在移除)
		// 通过PingReq超时断开连接 来触发断线重连
		if serviceNode := baseserver.GetServiceNode(msg.ServiceId); serviceNode == nil {
			// 添加连接上来的对端服务
			baseserver.AddServiceNode(iv.Session(), msg.ServiceId, msg.ServiceName, "remote")
			// 服务器之间的心跳检测
			// acceptor触发send connector触发rcv
			// 所以这里只能反应acceptor端的send和connector端的rcv是否正常
			iv.Session().HeartBeat(&socket.PingReq{})
		}
		return nil
	case *socket.PingReq:
		// 来自ServiceIdentifyACK接收端的服务器信息
		ctx := iv.Session().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(common.ContextSetCtxKey, &ed) {
				log.Printf("receive PingReq from [%v] session=%v \n", ed.Id, iv.Session().GetId())
			}
		}
		iv.Session().Send(&socket.PingAck{})
		return nil
	case *socket.PingAck:
		ctx := iv.Session().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(common.ContextSetCtxKey, &ed) {
				log.Printf("receive PingAck from [%v] session=%v \n", ed.Id, iv.Session().GetId())
			}
		}
		return nil
	case *socket.SessionClosed:
		sid := baseserver.RemoveServiceNode(iv.Session())
		log.Printf("SessionClosed sessionId=%v sid=%v \n", iv.Session().GetId(), sid)
		return nil
	default:
		log.Printf("receive unknown msg %v msgT:%v ivM %v \n", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
