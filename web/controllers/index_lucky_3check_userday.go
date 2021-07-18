package controllers

import (
	"fmt"
	"log"
	"lottery/conf"
	"lottery/models"
	"lottery/services"
	"lottery/web/utils"
	"strconv"
	"time"
)

//用户参与次数检测
//c *IndexController
func (api *LuckyApi) checkUserday(uid int, num int64) bool {
	//c.ServiceUserday.GetUserToday(uid)
	userdayService := services.NewUserdayService()
	//services.NewUserdayService().GetUserToday(uid)
	userdayInfo := userdayService.GetUserToday(uid)
	if userdayInfo != nil && userdayInfo.Uid == uid {
		//今天存在抽奖记录
		if userdayInfo.Num > conf.UserPrizeMax {
			if int(num) < userdayInfo.Num {
				//初始化一下，保证重启后数据库和缓存的用户累计次数一样
				utils.InitUserLuckyNum(uid, int64(userdayInfo.Num))
			}
			return false
		} else {
			userdayInfo.Num++
			if int(num) < userdayInfo.Num {
				//初始化一下，保证重启后数据库和缓存的用户累计次数一样
				utils.InitUserLuckyNum(uid, int64(userdayInfo.Num))
			}
			//c.ServiceUserday.Update(userdayInfo,nil)
			err103 := userdayService.Update(userdayInfo, nil)
			if err103 != nil {
				log.Println("index_lucky_check_userday ServiceUserDay.Update err103=", err103)
			}
		}
	} else {
		//创建今天的用户参与记录
		y, m, d := time.Now().Date()
		strDay := fmt.Sprintf("%d%02d%02d", y, m, d)
		day, _ := strconv.Atoi(strDay)
		userdayInfo = &models.LtUserday{
			Uid:        uid,
			Day:        day,
			Num:        1,
			SysCreated: int(time.Now().Unix()),
		}
		//c.ServiceUserday.Create(userdayInfo)
		err103 := userdayService.Create(userdayInfo)
		if err103 != nil {
			log.Println("index_lucky_check_userday ServiceUserDay.Create err103=", err103)
		}
		utils.InitUserLuckyNum(uid, 1)
	}
	return true
}
