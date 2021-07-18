package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/comm"
	"lottery/models"
	"lottery/services"
)

//后台ip地址黑名单管理
type AdminBlackipController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

// GET /admin/blackip/
func (c *AdminBlackipController) Get() mvc.Result {
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	datalist := c.ServiceBlackip.GetAll(page, size)
	total := (page-1)*size + len(datalist)
	// 数据总数
	if len(datalist) >= size {
		total = int(c.ServiceBlackip.CountAll())
		pageNext = fmt.Sprintf("%d", page+1)
	}
	if page > 1 {
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/blackip.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "blackip",
			"Datalist": datalist,
			"Total":    total,
			"Now":      comm.NowUnix(),
			"PagePrev": pagePrev,
			"PageNext": pageNext,
		},
		Layout: "admin/layout.html",
	}
}

//设置黑名单
// GET /admin/blackip/black?id=1&time=0
func (c *AdminBlackipController) GetBlack() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	t := c.Ctx.URLParamIntDefault("time", 0)
	if err == nil {
		//这里的t是天数
		if t > 0 {
			//用秒存进去
			t = t*86400 + comm.NowUnix()
		}
		c.ServiceBlackip.Update(&models.LtBlackip{Id: id, Blacktime: t, SysUpdated: comm.NowUnix()},
			[]string{"blacktime"})
	}
	return mvc.Response{
		Path: "/admin/blackip",
	}
}
