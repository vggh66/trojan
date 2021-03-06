package web

import (
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"github.com/unrolled/secure"
	"net/http"
	"strconv"
	"trojan/core"
	"trojan/util"
	"trojan/web/controller"
)

func userRouter(router *gin.Engine) {
	user := router.Group("/trojan/user")
	{
		user.GET("", func(c *gin.Context) {
			c.JSON(200, controller.UserList())
		})
		user.POST("", func(c *gin.Context) {
			username := c.PostForm("username")
			password := c.PostForm("password")
			c.JSON(200, controller.CreateUser(username, password))
		})
		user.DELETE("", func(c *gin.Context) {
			stringId := c.Query("id")
			id, _ := strconv.Atoi(stringId)
			c.JSON(200, controller.DelUser(uint(id)))
		})
	}
}

func dataRouter(router *gin.Engine) {
	data := router.Group("/trojan/data")
	{
		data.POST("", func(c *gin.Context) {
			sID := c.PostForm("id")
			sQuota := c.PostForm("quota")
			id, _ := strconv.Atoi(sID)
			quota, _ := strconv.Atoi(sQuota)
			c.JSON(200, controller.SetData(uint(id), quota))
		})
		data.DELETE("", func(c *gin.Context) {
			sID := c.Query("id")
			id, _ := strconv.Atoi(sID)
			c.JSON(200, controller.CleanData(uint(id)))
		})
	}
}

func commonRouter(router *gin.Engine) {
	common := router.Group("/common")
	{
		common.GET("/version", func(c *gin.Context) {
			c.JSON(200, controller.Version())
		})
		common.POST("/loginInfo", func(c *gin.Context) {
			c.JSON(200, controller.SetLoginInfo(c.PostForm("title")))
		})
	}
}

func staticRouter(router *gin.Engine) {
	box := packr.New("trojanBox", "./templates")
	router.Use(func(c *gin.Context) {
		requestUrl := c.Request.URL.Path
		if box.Has(requestUrl) || requestUrl == "/" {
			http.FileServer(box).ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	})
}

func sslRouter(router *gin.Engine, port int) {
	domain, _ := core.GetValue("domain")
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			secureMiddleware := secure.New(secure.Options{
				SSLRedirect: true,
				SSLHost:     fmt.Sprintf("%s:%d", domain, port),
			})
			err := secureMiddleware.Process(c.Writer, c.Request)
			// If there was an error, do not continue.
			if err != nil {
				return
			}

			c.Next()
		}
	}()
	router.Use(secureFunc)
}

// Start web启动入口
func Start(port int, isSSL bool) {
	router := gin.Default()
	if isSSL {
		sslRouter(router, port)
	}
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	staticRouter(router)
	router.Use(Auth(router).MiddlewareFunc())
	userRouter(router)
	dataRouter(router)
	commonRouter(router)
	util.OpenPort(port)
	if isSSL {
		config := core.Load("")
		ssl := &config.SSl
		util.OpenPort(80)
		go router.Run(":80")
		router.RunTLS(fmt.Sprintf(":%d", port), ssl.Cert, ssl.Key)
	} else {
		router.Run(fmt.Sprintf(":%d", port))
	}
}
