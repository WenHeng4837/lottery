package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/comm"
	"lottery/models"
	"lottery/services"
)

//后台用户管理
type AdminUserController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

// GET /admin/user/
func (c *AdminUserController) Get() mvc.Result {
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	datalist := c.ServiceUser.GetAll(page, size)
	total := (page-1)*size + len(datalist)
	// 数据总数
	if len(datalist) >= size {
		total = c.ServiceUser.CountAll()
		pageNext = fmt.Sprintf("%d", page+1)
	}
	if page > 1 {
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/user.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "user",
			"Datalist": datalist,
			"Total":    total,
			"Now":      comm.NowUnix(),
			"PagePrev": pagePrev,
			"PageNext": pageNext,
		},
		Layout: "admin/layout.html",
	}
}

// GET /admin/user/black?id=1&time=0
//设置黑名单或者洗白
func (c *AdminUserController) GetBlack() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	//这里的t是天数
	t := c.Ctx.URLParamIntDefault("time", 0)
	if err == nil {
		if t > 0 {
			//得变成秒存进去,从当前时间开始
			t = t*86400 + comm.NowUnix()
		}
		//blacktime强制更新
		c.ServiceUser.Update(&models.LtUser{Id: id, Blacktime: t, SysUpdated: comm.NowUnix()},
			[]string{"blacktime"})
	}
	return mvc.Response{
		Path: "/admin/user",
	}
}
