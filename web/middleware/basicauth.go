// file: middleware/basicauth.go

package middleware

import "github.com/kataras/iris/v12/middleware/basicauth"

// BasicAuth middleware sample.
//配置了一个后台管理员用户名和密码
var BasicAuth = basicauth.New(basicauth.Config{
	Users: map[string]string{
		"admin": "password",
	},
})
