package controllers

import (
	"fmt"
	"log"
	"lottery/services"

	"lottery/comm"
	"lottery/conf"
	"lottery/models"
	"lottery/web/utils"
)

type LuckyApi struct {
}

func (api *LuckyApi) luckyDo(uid int, username, ip string) (int, string, *models.ObjGiftPrize) {

	//2用户抽奖分布式锁定
	ok := utils.LockLucky(uid)
	if ok {
		defer utils.UnlockLucky(uid)
	} else {

		return 102, "正在抽奖，请稍后重试", nil
	}
	//3验证用户今日参与次数
	userDayNum := utils.IncrUserLuckyNum(uid)
	if userDayNum > conf.UserPrizeMax {

		return 103, "今日的抽奖次数已经用完，明天再来吧", nil
	} else {
		// c.checkUserday(uid,userDayNum)
		ok = api.checkUserday(uid, userDayNum)
		if !ok {

			return 103, "今日的抽奖次数已经用完，明天再来吧", nil
		}
	}
	//4验证IP今日的参与次数
	ipDayNum := utils.IncrIpLuckyNum(ip)
	if ipDayNum > conf.IpLimitMax {

		return 104, "相同IP参与次数太多，明天再来吧", nil
	}
	limitBlack := false //黑名单
	//如果大于最大数肯定立马黑名单
	if ipDayNum > conf.IpPrizeMax {
		limitBlack = true
	}
	//5验证IP黑名单
	var blackipInfo *models.LtBlackip
	if !limitBlack {
		ok, blackipInfo = api.checkBlackip(ip)
		if !ok {
			fmt.Println("黑名单中的IP", ip, limitBlack)
			limitBlack = true
		}
	}
	//6验证用户黑名单
	var userInfo *models.LtUser
	if !limitBlack {
		ok, userInfo = api.checkBlackUser(uid)
		if !ok {
			fmt.Println("黑名单中的用户", uid, limitBlack)
			limitBlack = true
		}
	}
	//7获得抽奖编码
	prizeCode := comm.Random(10000)
	fmt.Println("hahahaha:%d", prizeCode)
	//8匹配奖品是否中奖
	prizeGift := api.prize(prizeCode, limitBlack)
	if prizeGift == nil ||
		prizeGift.PrizeNum < 0 ||
		(prizeGift.PrizeNum > 0 && prizeGift.LeftNum <= 0) {

		return 205, "很遗憾，没有中奖，请下次再试", nil
	}
	//9有限制奖品发放
	if prizeGift.PrizeNum > 0 {
		//缓存没有奖品了
		if utils.GetGiftPoolNum(prizeGift.Id) <= 0 {

			return 206, "很遗憾，没有中奖，请下次再试", nil
		}
		ok = utils.PrizeGift(prizeGift.Id, prizeGift.LeftNum)
		if !ok {

			return 206, "很遗憾，没有中奖，请下次再试", nil
		}
	}

	//10不同编码优惠券发放
	if prizeGift.Gtype == conf.GtypeCodeDiff {
		//code := utils.PrizeCodeDiff(prizeGift.Id, c.ServiceCode)
		code := utils.PrizeCodeDiff(prizeGift.Id, services.NewCodeService())
		//code是虚拟卷编码，不是虚拟卷主键
		if code == "" {

			return 208, "很遗憾，没有中奖，请下次再试", nil
		}
		prizeGift.Gdata = code
	}
	//11记录中奖信息,这个不能修改
	result := models.LtResult{
		GiftId:     prizeGift.Id,
		GiftName:   prizeGift.Title,
		GiftType:   prizeGift.Gtype,
		Uid:        uid,
		Username:   username,
		PrizeCode:  prizeCode,
		GiftData:   prizeGift.Gdata,
		SysCreated: comm.NowUnix(),
		SysIp:      ip,
		SysStatus:  0,
	}
	//c.ServiceResult.Create(&result)
	err := services.NewResultService().Create(&result)
	if err != nil {
		log.Println("index_lucky.GetLucky ServiceResult.Create ", result,
			", error=", err)

		return 209, "很遗憾，没有中奖，请下次再试", nil
	}
	if prizeGift.Gtype == conf.GtypeGiftLarge {
		// 如果获得了实物大奖，需要将用户、IP设置成黑名单一段时间
		api.prizeLarge(ip, uid, username, userInfo, blackipInfo)
	}
	//12返回抽奖结果
	return 0, "", prizeGift
}
