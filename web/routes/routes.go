package routes

import (
	"github.com/kataras/iris/v12/mvc"
	"lottery/bootstrap"
	"lottery/services"
	"lottery/web/controllers"
	"lottery/web/middleware"
)

// Configure registers the necessary routes to the app.
func Configure(b *bootstrap.Bootstrapper) {
	userService := services.NewUserService()
	giftService := services.NewGiftService()
	codeService := services.NewCodeService()
	resultService := services.NewResultService()
	userdayService := services.NewUserdayService()
	blackipService := services.NewBlackipService()

	index := mvc.New(b.Party("/"))
	//全部注册一遍
	index.Register(userService, giftService, codeService, resultService, userdayService, blackipService)
	index.Handle(new(controllers.IndexController))

	//后台管理首页
	admin := mvc.New(b.Party("/admin"))
	//中间件
	admin.Router.Use(middleware.BasicAuth)
	admin.Register(userService, giftService, codeService, resultService, userdayService, blackipService)
	admin.Handle(new(controllers.AdminController))

	//后台用户黑名单管理
	adminUser := admin.Party("/user")
	adminUser.Register(userService)
	adminUser.Handle(new(controllers.AdminUserController))

	//后台奖品,这里不直接从mvc里new，而是从admin就可以继承到admin的路径以及中间件
	adminGift := admin.Party("/gift")
	adminGift.Register(giftService)
	adminGift.Handle(new(controllers.AdminGiftController))

	//后台优惠券管理
	adminCode := admin.Party("/code")
	adminCode.Register(codeService)
	adminCode.Handle(new(controllers.AdminCodeController))

	//后台中奖记录管理
	adminResult := admin.Party("/result")
	adminResult.Register(resultService)
	adminResult.Handle(new(controllers.AdminResultController))

	//后台地址黑名单管理
	adminBlackip := admin.Party("/blackip")
	adminBlackip.Register(blackipService)
	adminBlackip.Handle(new(controllers.AdminBlackipController))
	//
	rpc := mvc.New(b.Party("/rpc"))
	rpc.Register(userService, giftService, codeService, resultService, userdayService, blackipService)
	rpc.Handle(new(controllers.RpcController))
}
