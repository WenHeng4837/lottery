package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/comm"
	"lottery/conf"
	"lottery/models"
	"lottery/services"
	"lottery/web/utils"
	"strings"
)

//后台管理优惠券
type AdminCodeController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

func (c *AdminCodeController) Get() mvc.Result {
	giftId := c.Ctx.URLParamIntDefault("gift_id", 0)
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	var datalist []models.LtCode
	//数据库数量，缓存数量
	var num int
	var cacheNum int
	if giftId > 0 {
		//有传数据就查一条
		datalist = c.ServiceCode.Search(giftId)
		num, cacheNum = utils.GetCacheCodeNum(giftId, c.ServiceCode)
	} else {
		//没有就搜所有
		//页码，每页多少条
		datalist = c.ServiceCode.GetAll(page, size)
	}
	//偏移量再加上这次查出来的(如果当前查出来的数不足一页就直接在这里加上就好)
	total := (page-1)*size + len(datalist)
	// 数据总数
	//如果查出来的数与总数一样就有可能存在下一页,这种情况下total算出来不准的
	if len(datalist) >= size {
		if giftId > 0 {
			total = int(c.ServiceCode.CountByGift(giftId))
		} else {
			total = int(c.ServiceCode.CountAll())
		}
		//有上一页
		pageNext = fmt.Sprintf("%d", page+1)
	}
	//有下一页
	if page > 1 {
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/code.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "code",
			"GiftId":   giftId,
			"Datalist": datalist,
			"Total":    total,
			"PagePrev": pagePrev,
			"PageNext": pageNext,
			"CodeNum":  num,
			"CacheNum": cacheNum,
		},
		Layout: "admin/layout.html",
	}
}

//导入优惠券
func (c *AdminCodeController) PostImport() {
	giftId := c.Ctx.URLParamIntDefault("gift_id", 0)
	fmt.Println("PostImport giftId=", giftId)
	//参数有异常处理
	if giftId < 1 {
		c.Ctx.Text("没有指定奖品ID，无法进行导入，<a href='' onclick='history.go(-1);return false;'>返回</a>")
		return
	}
	gift := c.ServiceGift.Get(giftId, true)
	//虚拟券，不同的码conf.GtypeCodeDiff
	if gift == nil || gift.Gtype != conf.GtypeCodeDiff {
		c.Ctx.HTML("没有指定的优惠券类型的奖品，无法进行导入，<a href='' onclick='history.go(-1);return false;'>返回</a>")
		return
	}
	//从参数里获取优惠券的码
	codes := c.Ctx.PostValue("codes")
	//时间
	now := comm.NowUnix()
	//分割,一行一个数据
	list := strings.Split(codes, "\n")
	//导入成功数量、失败数量
	sucNum := 0
	errNum := 0
	for _, code := range list {
		//去掉一些多余空格
		code := strings.TrimSpace(code)
		if code != "" {
			data := &models.LtCode{
				GiftId:     giftId,
				Code:       code,
				SysCreated: now,
				SysUpdated: now,
			}
			//导入数据库
			err := c.ServiceCode.Create(data)
			if err != nil {
				errNum++
			} else {
				// 成功导入数据库，还需要导入到缓存中
				ok := utils.ImportCacheCodes(giftId, code)
				if ok {
					sucNum++
				} else {
					errNum++
				}
			}
		}
	}
	c.Ctx.HTML(fmt.Sprintf("成功导入 %d 条，导入失败 %d 条，<a href='/admin/code?gift_id=%d'>返回</a>", sucNum, errNum, giftId))
}

//删除
func (c *AdminCodeController) GetDelete() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceCode.Delete(id)
	}
	//做一次页面刷新
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/code"
	}
	return mvc.Response{
		Path: refer,
	}
}

//重置
func (c *AdminCodeController) GetReset() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	now := comm.NowUnix()
	if err == nil {
		c.ServiceCode.Update(&models.LtCode{Id: id, SysUpdated: now, SysStatus: 0}, []string{"sys_status"})
	}
	//做一次页面刷新
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/code"
	}
	return mvc.Response{
		Path: refer,
	}
}

// 重新整理优惠券的数据，如果是本地服务，也需要启动时加载
func (c *AdminCodeController) GetRecache() {
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/code"
	}
	id, err := c.Ctx.URLParamInt("id")
	if id < 1 || err != nil {
		rs := fmt.Sprintf("没有指定优惠券所属的奖品id, <a href='%s'>返回</a>", refer)
		//输出
		c.Ctx.HTML(rs)
		return
	}
	sucNum, errNum := utils.RecacheCodes(id, c.ServiceCode)

	rs := fmt.Sprintf("sucNum=%d, errNum=%d, <a href='%s'>返回</a>", sucNum, errNum, refer)
	c.Ctx.HTML(rs)
}
