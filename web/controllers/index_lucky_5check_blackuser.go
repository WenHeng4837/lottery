package controllers

import (
	"lottery/models"
	"lottery/services"
	"time"
)

///用户黑名单检测
//c *IndexController
func (api *LuckyApi) checkBlackUser(uid int) (bool, *models.LtUser) {
	//c.ServiceUser.Get(uid)
	info := services.NewUserService().Get(uid)
	if info != nil && info.Blacktime > int(time.Now().Unix()) {
		// 黑名单存在并且有效
		return false, info
	}
	return true, info
}
