package handler

import (
	dao "filestore-byceph/db"
	//"encoding/json"
	"filestore-byceph/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	userPasswordSalt = "!#FDGS@#%"
)

//SignUpHandler:响应注册页面
func SignUpHandler(c *gin.Context) () {

	c.Redirect(http.StatusFound, "/static/view/signup.html")

}

//DoSignUpHandler:处理注册请求
func DoSignUpHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("passwd")

	if len(username) < 3 || len(passwd) < 5{
		c.JSON(http.StatusOK, gin.H{
			"msg": "Invaild Parameter",
			"code": -1,
		})
		return
	}

	encode_passwd := utils.Sha1([]byte(passwd + userPasswordSalt))
	//将用户信息注册到用户表中
	ret := dao.UserSignUp(username, encode_passwd)
	if ret {
		c.JSON(http.StatusOK, gin.H{
			"msg": "SignUp Succeeded",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "SignUp Failed",
			"code": 0,
		})
	}
}
//SignInHandler：返回登录响应页面
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

//DoSignInHandler：登录接口
func DoSignInHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("passwd")
	//todo:校验输出参数

	//加密用户输入的密码
	encodePassword := utils.Sha1([]byte(passwd+userPasswordSalt))


	//1.校验用户名密码
	passwdChecked := dao.UserSignIn(username, encodePassword)

	if !passwdChecked {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg": "密码错误",
		})
		return
	}

	//2.生成访问的token
	token := GenToken(username)
	updateTokenResult := dao.UpdateUserToken(username, token)
	if !updateTokenResult {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg": "更新token失败",
		})
		return
	}
	resp := utils.RespMsg{
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
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())

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
func UserInfoHandler(c *gin.Context) () {
	username := c.Request.FormValue("username")
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg": "查询不到信息",
		})
		return
	}

	//返回相应数据
	resp := utils.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())

}

func ValidToken(token string) bool {
	//判断token的时效性
	//数据库中比取出username对应的token
	//两个token进行对比
	return true
}