package controllers

import (
	"lottery/models"
	"lottery/services"
	"time"
)

//ip黑名单检测
//c *IndexController
func (api *LuckyApi) checkBlackip(ip string) (bool, *models.LtBlackip) {
	//c.ServiceBlackip.GetByIp(ip)
	info := services.NewBlackipService().GetByIp(ip)
	if info == nil || info.Ip == "" {
		return true, nil
	}
	if info.Blacktime > int(time.Now().Unix()) {
		// IP黑名单存在，并且还在黑名单有效期内
		return false, info
	}
	return true, info
}
