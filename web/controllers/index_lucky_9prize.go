package controllers

import (
	"lottery/conf"
	"lottery/models"
	"lottery/services"
)

//后台抽奖
//c *IndexController
func (api *LuckyApi) prize(prizeCode int, limitBlack bool) *models.ObjGiftPrize {
	var prizeGift *models.ObjGiftPrize
	//获取可用的奖品
	//c.ServiceGift.GetAllUse(true)
	giftList := services.NewGiftService().GetAllUse(true)
	for _, gift := range giftList {
		if gift.PrizeCodeA <= prizeCode && gift.PrizeCodeB >= prizeCode {
			// 中奖编码区间满足条件，说明可以中奖
			//不是黑名单或者不会是实物奖就直接返回，不然不能直接返回
			if !limitBlack || gift.Gtype < conf.GtypeGiftSmall {
				//fmt.Println(gift)
				prizeGift = &gift
				break
			}
		}
	}
	return prizeGift
}
