package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/comm"
	"lottery/models"
	"lottery/services"
	"lottery/web/utils"
	"lottery/web/viewmodels"
	"time"
)

//后台管理奖品
type AdminGiftController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

//显示获取奖品
func (c *AdminGiftController) Get() mvc.Result {
	// 数据列表
	datalist := c.ServiceGift.GetAll(false)
	//做一个循环处理,把需要返回給前端页面的数据转换一下格式
	for i, giftInfo := range datalist {
		// 奖品发放的计划数据
		//反序列化处理,[]byte(giftInfo.PrizeData)转成byte切片然后填充进去
		prizedata := make([][2]int, 0)
		err := json.Unmarshal([]byte(giftInfo.PrizeData), &prizedata)
		if err != nil || prizedata == nil || len(prizedata) < 1 {
			datalist[i].PrizeData = "[]"
		} else {
			newpd := make([]string, len(prizedata))
			for index, pd := range prizedata {
				//格式化时间戳
				ct := comm.FormatFromUnixTime(int64(pd[0]))
				//格式化成一个字符串放进去
				//在数据库的计划任务每这个字段存的就是时间，数量
				newpd[index] = fmt.Sprintf("【%s】: %d", ct, pd[1])
			}
			//在序列化一下
			str, err := json.Marshal(newpd)
			if err == nil && len(str) > 0 {
				datalist[i].PrizeData = string(str)
			} else {
				datalist[i].PrizeData = "[]"
			}
		}
		// 奖品当前的奖品池数量
		num := utils.GetGiftPoolNum(giftInfo.Id)
		datalist[i].Title = fmt.Sprintf("【%d】%s", num, datalist[i].Title)
	}
	total := len(datalist)
	return mvc.View{
		Name: "admin/gift.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "gift",
			"Datalist": datalist,
			"Total":    total,
		},
		Layout: "admin/layout.html",
	}
}

//编辑奖品,编辑后点板寸那个按钮调用的是下面postsave函数，这个函数会修改更新时间
func (c *AdminGiftController) GetEdit() mvc.Result {
	//设个默认值为0
	id := c.Ctx.URLParamIntDefault("id", 0)
	giftInfo := viewmodels.ViewGift{}
	if id > 0 {
		data := c.ServiceGift.Get(id, false)
		if data != nil {
			giftInfo.Id = data.Id
			giftInfo.Title = data.Title
			giftInfo.PrizeNum = data.PrizeNum
			giftInfo.PrizeCode = data.PrizeCode
			giftInfo.PrizeTime = data.PrizeTime
			giftInfo.Img = data.Img
			giftInfo.Displayorder = data.Displayorder
			giftInfo.Gtype = data.Gtype
			giftInfo.Gdata = data.Gdata
			//时间转换成整型
			giftInfo.TimeBegin = comm.FormatFromUnixTime(int64(data.TimeBegin))
			giftInfo.TimeEnd = comm.FormatFromUnixTime(int64(data.TimeEnd))
		}
	}
	return mvc.View{
		Name: "admin/giftEdit.html",
		Data: iris.Map{
			"Title":   "管理后台",
			"Channel": "gift",
			"info":    giftInfo,
		},
		Layout: "admin/layout.html",
	}
}

//数据更新保存
func (c *AdminGiftController) PostSave() mvc.Result {
	data := viewmodels.ViewGift{}
	err := c.Ctx.ReadForm(&data)

	if err != nil {
		fmt.Println("admin_gift.PostSave ReadForm error=", err)
		return mvc.Response{
			Text: fmt.Sprintf("ReadForm转换异常, err=%s", err),
		}
	}
	giftInfo := models.LtGift{}
	giftInfo.Id = data.Id
	giftInfo.Title = data.Title
	giftInfo.PrizeNum = data.PrizeNum
	giftInfo.PrizeCode = data.PrizeCode
	giftInfo.PrizeTime = data.PrizeTime
	giftInfo.Img = data.Img
	giftInfo.Displayorder = data.Displayorder
	giftInfo.Gtype = data.Gtype
	giftInfo.Gdata = data.Gdata
	t1, err1 := comm.ParseTime(data.TimeBegin)
	t2, err2 := comm.ParseTime(data.TimeEnd)
	if err1 != nil || err2 != nil {
		return mvc.Response{
			Text: fmt.Sprintf("开始时间、结束时间的格式不正确, err1=%s, err2=%s", err1, err2),
		}
	}
	giftInfo.TimeBegin = int(t1.Unix())
	giftInfo.TimeEnd = int(t2.Unix())
	if giftInfo.Id > 0 {
		//先把原来数据拿出来对比下，存在才是数据更新否则就收添加
		datainfo := c.ServiceGift.Get(giftInfo.Id, false)
		if datainfo != nil {
			//修改时间更新
			giftInfo.SysUpdated = int(time.Now().Unix())
			giftInfo.SysIp = comm.ClientIP(c.Ctx.Request())
			// 对比修改的内容项
			if datainfo.PrizeNum != giftInfo.PrizeNum {
				// 奖品总数量发生了改变
				giftInfo.LeftNum = datainfo.LeftNum - (datainfo.PrizeNum - giftInfo.PrizeNum)
				//减完小于0重置
				if giftInfo.LeftNum < 0 || giftInfo.PrizeNum <= 0 {
					giftInfo.LeftNum = 0
				}
				giftInfo.SysStatus = datainfo.SysStatus
				////数量发生变化，更新奖品的发奖计划
				utils.ResetGiftPrizeData(&giftInfo, c.ServiceGift)
			}
			if datainfo.PrizeTime != giftInfo.PrizeTime {
				// 发奖周期发生了变化，更新奖品的发奖计划
				utils.ResetGiftPrizeData(&giftInfo, c.ServiceGift)
			}
			c.ServiceGift.Update(&giftInfo, []string{"title", "prize_num", "left_num", "prize_code", "prize_time", "img", "displayorder", "gtype", "gdata", "time_begin", "time_end", "sys_updated"})
		} else {
			giftInfo.Id = 0
		}
	}
	//其实就是如果没有数据这里就插入一条新记录
	if giftInfo.Id <= 0 {
		giftInfo.LeftNum = giftInfo.PrizeNum
		giftInfo.SysIp = comm.ClientIP(c.Ctx.Request())
		giftInfo.SysCreated = int(time.Now().Unix())
		c.ServiceGift.Create(&giftInfo)

		// 更新奖品的发奖计划
		utils.ResetGiftPrizeData(&giftInfo, c.ServiceGift)
	}
	//做个跳转
	return mvc.Response{
		Path: "/admin/gift",
	}
}

//数据删除
func (c *AdminGiftController) GetDelete() mvc.Result {

	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceGift.Delete(id)
	}
	return mvc.Response{
		Path: "/admin/gift",
	}
}

//数据重置，也就是恢复
func (c *AdminGiftController) GetReset() mvc.Result {

	id, err := c.Ctx.URLParamInt("id")
	now := comm.NowUnix()
	if err == nil {
		//sys_status即使为空也要更新
		c.ServiceGift.Update(&models.LtGift{Id: id, SysStatus: 0, SysUpdated: now}, []string{"sys_status"})
	}
	return mvc.Response{
		Path: "/admin/gift",
	}
}
