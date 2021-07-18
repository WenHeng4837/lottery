package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/services"
)

type AdminController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

//后台管理首页，返回值是mvc模板
func (c *AdminController) Get() mvc.Result {
	return mvc.View{
		Name: "admin/index.html",
		Data: iris.Map{
			"Title":   "管理后台",
			"Channel": "",
		},
		Layout: "admin/layout.html",
	}
}
