package handler

import (
	"filestore-byceph/service/account/proto"
	"github.com/micro/go-micro"
)

var (
	userCli proto.UserService
)

func init()  {
	service := micro.NewService(  //apiGateway并不注册到注册中心，可以不用使用Name方法
			micro.Name("go.micro.api.user"))
	//可以再初始化中解析注册中心等初始化工作
	service.Init()

	//初始化一个rpcClient
	//userClient = proto.NewUserService("go.micro.service.user",service.Client())
}