package main

import (
	"common"
	"common/service"
	"common/socket"
	_ "common/socket/tcp"
	"log"
)

func main() {
	log.Println("server starting ...")
	node := socket.NewServerNode(common.SocketTypTcpAcceptor, "test", "0.0.0.0:2701")
	node.Start()
	log.Println("server start success")
	service.WaitExitSignal()
	service.Stop(node)
	log.Println("server close")
}
