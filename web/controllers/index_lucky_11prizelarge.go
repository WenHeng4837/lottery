package controllers

import (
	"lottery/comm"
	"lottery/models"
	"lottery/services"
)

//实物大奖用户需要黑名单处理一阵子
//c *IndexController
func (api *LuckyApi) prizeLarge(ip string, uid int, username string, userinfo *models.LtUser, blackipInfo *models.LtBlackip) {
	userService := services.NewUserService()
	blackipService := services.NewBlackipService()
	//当前时间
	nowTime := comm.NowUnix()
	//黑名单时间
	blackTime := 30 * 86400
	// 更新用户的黑名单信息
	if userinfo == nil || userinfo.Id <= 0 || userinfo.Username == "" {
		userinfo = &models.LtUser{
			Id:         uid,
			Username:   username,
			Blacktime:  nowTime + blackTime,
			SysCreated: nowTime,
			SysIp:      ip,
		}
		//c.ServiceUser.Create(userinfo)
		userService.Create(userinfo)
	} else {
		userinfo = &models.LtUser{
			Id:         uid,
			Blacktime:  nowTime + blackTime,
			SysUpdated: nowTime,
		}
		//c.ServiceUser.Update(userinfo, nil)
		userService.Update(userinfo, nil)
	}
	// 更新要IP的黑名单信息
	if blackipInfo == nil || blackipInfo.Id <= 0 {
		blackipInfo = &models.LtBlackip{
			Ip:         ip,
			Blacktime:  nowTime + blackTime,
			SysCreated: nowTime,
		}
		//c.ServiceBlackip.Create(blackipInfo)
		blackipService.Create(blackipInfo)
	} else {
		//有了这个黑名单了修改黑名单时间
		blackipInfo.Blacktime = nowTime + blackTime
		blackipInfo.SysUpdated = nowTime
		//c.ServiceBlackip.Update(blackipInfo, nil)
		blackipService.Update(blackipInfo, nil)
	}
}
