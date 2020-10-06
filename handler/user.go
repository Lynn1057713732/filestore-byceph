package handler

import (
	//"encoding/json"
	"filestore-byceph/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	dao "filestore-byceph/db"
	"time"
)

const (
	userPasswordSalt = "!#FDGS@#%"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request) () {
	if r.Method == http.MethodGet{
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return

	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("passwd")

	if len(username) < 3 || len(passwd) < 5{
		w.Write([]byte("Invaild Parameter"))
		return
	}

	encode_passwd := utils.Sha1([]byte(passwd + userPasswordSalt))
	ret := dao.UserSignUp(username, encode_passwd)
	if ret {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Success"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed"))
	}
}

//SignInHandler：登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) () {
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("passwd")
	//todo:校验输出参数

	//加密用户输入的密码
	encodePassword := utils.Sha1([]byte(passwd+userPasswordSalt))


	//1.校验用户名密码
	passwdChecked := dao.UserSignIn(username, encodePassword)

	if !passwdChecked {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("密码错误"))
		return
	}

	//2.生成访问的token
	token := GenToken(username)
	updateTokenResult := dao.UpdateUserToken(username, token)
	if !updateTokenResult {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("更新token失败"))
		return
	}

	response := utils.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: struct {
			Username string
			Token string
		}{
			Username: username,
			Token: token,
		},
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response.JSONBytes())

}

//GenToken:生成token
func GenToken(username string) string{
	//40位的token
	//md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := utils.MD5([]byte(username + ts + "_tokenSalt"))
	return tokenPrefix + ts[:8]
}

//UserInfoHandler:返回用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) () {
	r.ParseForm()
	username := r.Form.Get("username")
	//token := r.Form.Get("token")
	//
	////验证token是否有效
	//isValidToken := ValidToken(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//查询用户信息
	user, err := dao.GetUserInfoByToken(username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//返回相应数据
	resp := utils.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: user,
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp.JSONBytes())
}

func ValidToken(token string) bool {
	//判断token的时效性
	//数据库中比取出username对应的token
	//两个token进行对比
	return true
}