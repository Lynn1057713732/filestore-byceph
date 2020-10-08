package route

import (
	"filestore-byceph/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine{
	// gin framework, 包括Logger, Recovery
	router := gin.Default()

	// 处理静态资源
	router.Static("/static/", "./static")

	// // 加入中间件，用于校验token的拦截器(将会从account微服务中验证)
	// router.Use(handler.HTTPInterceptor())

	// 使用gin插件支持跨域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS", "PATCH", "DELETE"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))

	//不需要用户token校验的
	router.GET("/user/signup", handler.SignUpHandler)
	router.POST("/user/signup", handler.DoSignUpHandler)

	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	//加入中间件，用户token校验
	// Use之后的所有handler都会经过拦截器进行token校验
	router.Use(handler.HTTPInterceptor())

	//文件存取接口
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/upload/success", handler.UploadSuccessHandler)
	router.GET("/file/meta", handler.GetFileMetaHandler)
	router.GET("/file/query", handler.FileQueryHandler)
	router.POST("/file/download", handler.DownloadHandler)

	router.PATCH("/file/update", handler.FileMetaUpdateHandler)
	router.DELETE("/file/delete", handler.FileDeleteHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)

	//秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	//分块上传接口
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppar", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)
	router.POST("/file/mpupload/cancel", handler.CancelUploadPartHandler)
	router.POST("/file/mpupload/status", handler.MultipartUploadStatusHandler)


	return router

}

