package service

import (
	"common"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/socket"
	_ "common/socket/tcp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type NetNodeParam struct {
	ServerTyp            string
	ServerName           string
	Addr                 string
	Typ                  int
	Zone                 int
	Index                int
	DiscoveryServiceName string // 用于服务发现
}

// CreateAcceptor 创建监听节点
func CreateAcceptor(param NetNodeParam) iface.INetNode {
	node := socket.NewServerNode(param.ServerTyp, param.ServerName, param.Addr)
	node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
	msgHandle := GetMsgHandle(0)
	node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)

	property := node.(common.ServerNodeProperty)
	property.SetServerTyp(param.Typ)
	property.SetZone(param.Zone)
	property.SetIndex(param.Index)

	node.Start()
	plugins.ETCDRegister(node)
	return node
}

// CreateConnector 创建连接节点
func CreateConnector(param NetNodeParam) iface.INetNode {
	plugins.DiscoveryService(param.DiscoveryServiceName, param.Zone,
		func(mn plugins.MultiServerNode, ed *plugins.ETCDServiceDesc) {
			// 不连接自己
			if ed.Typ == param.Typ && ed.Zone == param.Zone && ed.Index == param.Index {
				return
			}
			node := socket.NewServerNode(param.ServerTyp, param.ServerName, ed.Host)
			msgHandle := GetMsgHandle(0)
			node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
			node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)
			node.Start()
		})
	return nil
}

func Init() {
	err := plugins.InitServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		log.Println("InitServiceDiscovery err:", err)
		return
	}
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}

func Stop(node iface.INetNode) {
	if node == nil {
		return
	}
	node.Stop()
}
