package handler

import (
	"filestore-byceph/common"
	cfg "filestore-byceph/config"
	dao "filestore-byceph/db"
	"filestore-byceph/service/account/proto"
	"context"
	"filestore-byceph/utils"

)


type User struct {

}

//SignUp：处理用户注册请求
func (user *User)SignUp(ctx context.Context, req *proto.RequestSignUp, resp *proto.ResponseSignUp) error {
	username := req.Username
	passwd := req.Password

	if len(username) < 3 || len(passwd) < 5{
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return  nil
	}

	encodePassword := utils.Sha1([]byte(passwd + cfg.PasswordSalt))
	//将用户信息注册到用户表中
	ret := dao.UserSignUp(username, encodePassword)
	if ret {
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	} else {
		resp.Code = common.StatusRegisterFailed
		resp.Message = "注册失败"
	}
	return nil

}
