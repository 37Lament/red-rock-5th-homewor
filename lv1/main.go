package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// Binding from JSON
type Login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
//鉴权中间件描写
func m1(c*gin.Context)  {
	//获取cookie
	value, err := c.Cookie("gin_cookie")
	if err != nil {
		c.JSON(403,gin.H{
			"message":"认证失败,没有cookie",
		})
		//认证失败
		c.Abort()
	}else{
		//将获取到的cookie的值写入上下文
		c.Set("cookie",value)
		c.Next()
		v,_:=c.Get("next")
		fmt.Println(v)
	}
	
}
func main() {
	r:= gin.Default()

	r.POST("/login",func (c *gin.Context){
		username:=c.PostForm("username")
		password:=c.PostForm("password")
		if username == "123" && password == "321"{
			c.SetCookie("gin_cookie", username, 3600, "/", "", false, true)
			c.JSON(200,gin.H{
				"msg": "认证成功，已成功登录",
			})
		}else{
			c.JSON(403,gin.H{
				"message":"认证失败,账号密码错误",
			})
		}
	})
	//在中间放入鉴权中间件
	r.GET("/hello",m1, func(c *gin.Context) {
		cookie,_:=c.Get("cookie")
		str:=cookie.(string)
		c.String(200,"hello world"+str)
		//成功后弹出欢迎界面
	})
	r.Run()
}