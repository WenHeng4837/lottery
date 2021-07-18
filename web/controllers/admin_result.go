package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/models"
	"lottery/services"
)

//后台中奖记录
type AdminResultController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

//获取获奖记录
func (c *AdminResultController) Get() mvc.Result {
	giftId := c.Ctx.URLParamIntDefault("gift_id", 0)
	uid := c.Ctx.URLParamIntDefault("uid", 0)
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	var datalist []models.LtResult
	if giftId > 0 {
		//如果传了giftid
		datalist = c.ServiceResult.SearchByGift(giftId, page, size)
	} else if uid > 0 {
		//如果传了uid
		datalist = c.ServiceResult.SearchByUser(uid, page, size)
	} else {
		//如果参数都没有
		datalist = c.ServiceResult.GetAll(page, size)
	}
	total := (page-1)*size + len(datalist)
	// 数据总数
	if len(datalist) >= size {
		//如果传了giftid
		if giftId > 0 {
			total = int(c.ServiceResult.CountByGift(giftId))
		} else if uid > 0 {
			//如果传了uid
			total = int(c.ServiceResult.CountByUser(uid))
		} else {
			//如果参数都没有
			total = int(c.ServiceResult.CountAll())
		}
		//下一页
		pageNext = fmt.Sprintf("%d", page+1)
	}
	if page > 1 {
		//上一页
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/result.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "result",
			"GiftId":   giftId,
			"Uid":      uid,
			"Datalist": datalist,
			"Total":    total,
			"PagePrev": pagePrev,
			"PageNext": pageNext,
		},
		Layout: "admin/layout.html",
	}
}

//删除
func (c *AdminResultController) GetDelete() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceResult.Delete(id)
	}
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/result"
	}
	return mvc.Response{
		Path: refer,
	}
}

//设置作弊
func (c *AdminResultController) GetCheat() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		//状态为2就代表作弊
		c.ServiceResult.Update(&models.LtResult{Id: id, SysStatus: 2}, []string{"sys_status"})
	}
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/result"
	}
	return mvc.Response{
		Path: refer,
	}
}

//重置数据
func (c *AdminResultController) GetReset() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceResult.Update(&models.LtResult{Id: id, SysStatus: 0}, []string{"sys_status"})
	}
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/result"
	}
	return mvc.Response{
		Path: refer,
	}
}
