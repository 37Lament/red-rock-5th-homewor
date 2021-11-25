package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

// Binding from JSON

const filePath = "./users.data"

type User struct {
	Username string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type userHash map[string]string

type Checker struct {
	uh            userHash // 用户信息
	registerUsers []User   // 注册了但未保存的用户
}

func (c *Checker) SignIn(username, password string) {

	if _, ok := c.uh[username]; !ok {
		opt = 1
		return
	} else if c.uh[username] != password {
		opt = 2
		return
	} else {
		opt = 3
		return
	}
}

func (c *Checker) SignUp(username, password string) {
	if _, ok := c.uh[username]; ok {
		opt = 4
		return
	} else if len(password) < 6 {
		opt = 5
		return
	}
	{
		opt = 6
		c.registerUsers = append(c.registerUsers, User{
			Username: username,
			Password: password,
		})
		if len(c.registerUsers) > 10 {
			go c.Save()
		}
		c.uh[username] = password
		// 先写入缓存，再异步写入文件
		c.registerUsers = append(c.registerUsers, User{
			Username: username,
			Password: password,
		})
		if len(c.registerUsers) > 10 {
			go c.Save()
		}
		c.uh[username] = password
		return
	}}

func (c *Checker) Save() {
	fail := saveUsers(c.registerUsers)
	c.registerUsers = fail
}

func initUsers() (userHash, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer f.Close()

	uh := make(userHash)
	reader := bufio.NewReader(f)
	for {
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return nil, err
		}
		var user User
		err = json.Unmarshal(buf, &user)
		if err != nil {
			fmt.Println(err)
			continue
		}
		uh[user.Username] = user.Password
	}
	return uh, nil
}

func saveUsers(users []User) (fail []User) {
	// 以追加的方式写入文件
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, user := range users {
		buf, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
			fail = append(fail, user)
			continue
		}
		n, err := writer.Write(append(buf, byte('\n')))
		if err != nil {
			fmt.Println(n, err)
			fail = append(fail, user)
			continue
		}
	}
	writer.Flush()
	return
}

//鉴权中间件描写
func m1(c *gin.Context) {
	//获取cookie
	value, err := c.Cookie("gin_cookie")
	if err != nil {
		c.JSON(403, gin.H{
			"message": "认证失败,没有cookie",
		})
		//认证失败
		c.Abort()
	} else {
		//将获取到的cookie的值写入上下文
		c.Set("cookie", value)
		c.Next()
		v, _ := c.Get("next")
		fmt.Println(v)
	}

}

var opt int

func main() {
	r := gin.Default()
	checker := Checker{}
	var err error
	checker.uh, err = initUsers()
	if err != nil {
		return
	}
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		checker.SignIn(username, password)
		switch opt {
		case 1:
			c.JSON(403, gin.H{
				"message": "查无此人"})
		case 2:
			c.JSON(403, gin.H{
				"message": "认证失败,账号密码错误",
			})
		case 3:
			c.SetCookie("gin_cookie", username, 3600, "/", "", false, true)
			c.JSON(200, gin.H{
				"msg": "认证成功，已成功登录",
			})
		}
	})
	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		checker.SignUp(username, password)
		switch opt {
		case 4:
			c.JSON(403, gin.H{
				"msg": "用户名已被占用",
			})
		case 5:
			c.JSON(403, gin.H{
				"message": "密码长度应大于六位，请重新输入",
			})
		case 6:
			c.JSON(200, gin.H{
				"message": "注册成功",
			})
		}
	})

	//在中间放入鉴权中间件
	r.GET("/hello", m1, func(c *gin.Context) {
		cookie, _ := c.Get("cookie")
		str := cookie.(string)
		c.String(200, "hello world"+str)
		//成功后弹出欢迎界面
	})
	r.Run()
}
