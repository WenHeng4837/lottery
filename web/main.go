package main

import (
	"fmt"
	"lottery/bootstrap"
	"lottery/web/middleware/identity"
	"lottery/web/routes"
)

//端口
var port = 8080

func newApp() *bootstrap.Bootstrapper {
	//初始化应用
	app := bootstrap.New("Go抽奖系统", "wulihui")
	app.Bootstrap()
	app.Configure(identity.Configure, routes.Configure)
	return app
}

//启动类
func main() {
	app := newApp()
	app.Listen(fmt.Sprintf(":%d", port))
}
