package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"lottery/conf"
	"time"

	"github.com/gomodule/redigo/redis"
	"lottery/comm"
	"lottery/datasource"
	"lottery/models"
	"lottery/services"
)

func init() {
	// 本地开发测试的时候，每次重新启动，奖品池自动归零
	resetServGiftPool()
}

//
//重置一个奖品的发奖周期信息
//奖品剩余数量也会重新设置为当前奖品数量
//奖品的奖品池有效数量则会设置为空
//奖品数量、发放周期等设置有修改的时候，也需要重置
//【难点】根据发奖周期，重新更新发奖计划
func ResetGiftPrizeData(giftInfo *models.LtGift, giftService services.GiftService) {
	if giftInfo == nil || giftInfo.Id < 1 {
		return
	}
	id := giftInfo.Id
	nowTime := comm.NowUnix()
	// 不能发奖，不需要设置发奖周期
	if giftInfo.SysStatus == 1 || // 状态不对
		giftInfo.TimeBegin >= nowTime || // 开始时间不对
		giftInfo.TimeEnd <= nowTime || // 结束时间不对
		giftInfo.LeftNum <= 0 || // 剩余数不足
		giftInfo.PrizeNum <= 0 { // 总数不限制
		if giftInfo.PrizeData != "" {
			//清空旧的的发奖计划
			clearGiftPrizeData(giftInfo, giftService)
		}
		return
	}
	// 不限制发奖周期，直接把奖品数量全部设置上
	dayNum := giftInfo.PrizeTime
	if dayNum <= 0 {
		setGiftPool(id, giftInfo.LeftNum)
		return
	}

	// 重新计算出来合适的奖品发放节奏
	// 奖品池的剩余数先设置为空
	setGiftPool(id, 0)

	// 每天的概率一样
	// 一天内24小时，每个小时的概率是不一样的
	// 一小时内60分钟的概率一样
	prizeNum := giftInfo.PrizeNum
	//每天的数
	avgNum := prizeNum / dayNum

	// 每天可以分配到的奖品数量
	dayPrizeNum := make(map[int]int)
	// 平均分配，每天分到的奖品数量做分布
	if avgNum >= 1 && dayNum > 0 {
		for day := 0; day < dayNum; day++ {
			dayPrizeNum[day] = avgNum
		}
	}
	// 剩下的随机分配到任意哪天
	prizeNum -= dayNum * avgNum
	for prizeNum > 0 {
		prizeNum--
		day := comm.Random(dayNum)
		_, ok := dayPrizeNum[day]
		if !ok {
			dayPrizeNum[day] = 1
		} else {
			dayPrizeNum[day] += 1
		}
	}
	// 每天的map，每小时的map，60分钟的数组，奖品数量
	prizeData := make(map[int]map[int][60]int)
	for day, num := range dayPrizeNum {
		//map[int][60]int，结构：每小时每分钟数量
		dayPrizeData := getGiftPrizeDataOneDay(num)
		//结构：每天每小时每分钟数量
		prizeData[day] = dayPrizeData
	}
	// 将周期内每天、每小时、每分钟的数据 prizeData(发奖计划) 格式化([][2]int:时间、数量)，再序列化保存到数据表
	datalist := formatGiftPrizeData(nowTime, dayNum, prizeData)
	str, err := json.Marshal(datalist)
	if err != nil {
		log.Println("prizedata.ResetGiftPrizeData json error=", err)
	} else {
		// 保存奖品的分布计划数据
		info := &models.LtGift{
			Id:         giftInfo.Id,
			LeftNum:    giftInfo.PrizeNum,
			PrizeData:  string(str),
			PrizeBegin: nowTime,
			PrizeEnd:   nowTime + dayNum*86400,
			SysUpdated: nowTime,
		}
		err := giftService.Update(info, nil)
		if err != nil {
			log.Println("prizedata.ResetGiftPrizeData giftService.Update",
				info, ", error=", err)
		}
	}
}

/**
 * 根据奖品的发奖计划，把设定的奖品数量放入奖品池
 * 需要每分钟执行一次
 * 【难点】定时程序，根据奖品设置的数据，更新奖品池的数据
 */
func DistributionGiftPool() int {
	totalNum := 0
	now := comm.NowUnix()
	giftService := services.NewGiftService()
	//不要从缓存里读
	list := giftService.GetAll(false)
	if list != nil && len(list) > 0 {
		for _, gift := range list {
			// 是否正常状态
			if gift.SysStatus != 0 {
				continue
			}
			// 是否限量产品
			if gift.PrizeNum < 1 {
				continue
			}
			// 时间段是否正常
			if gift.TimeBegin > now || gift.TimeEnd < now {
				continue
			}
			// 计划数据的长度太短，不需要解析和执行
			// 发奖计划，[[时间1,数量1],[时间2,数量2]]
			if len(gift.PrizeData) <= 7 {
				continue
			}
			var cronData [][2]int
			err := json.Unmarshal([]byte(gift.PrizeData), &cronData)
			if err != nil {
				log.Println("prizedata.DistributionGiftPool Unmarshal error=", err)
			} else {
				index := 0
				giftNum := 0
				for i, data := range cronData {
					ct := data[0]
					num := data[1]
					if ct <= now {
						// 之前没有执行的数量，都要放进奖品池
						giftNum += num
						index = i + 1
					} else {
						break
					}
				}
				// 有奖品需要放入到奖品池
				if giftNum > 0 {
					incrGiftPool(gift.Id, giftNum)
					totalNum += giftNum
				}
				// 有计划数据被执行过，需要更新到数据库，更新发奖计划
				if index > 0 {
					if index >= len(cronData) {
						cronData = make([][2]int, 0)
					} else {
						cronData = cronData[index:]
					}
					// 更新到数据库
					str, err := json.Marshal(cronData)
					if err != nil {
						log.Println("prizedata.DistributionGiftPool Marshal(cronData)", cronData, "error=", err)
					}
					columns := []string{"prize_data"}
					err = giftService.Update(&models.LtGift{
						Id:        gift.Id,
						PrizeData: string(str),
					}, columns)
					if err != nil {
						log.Println("prizedata.DistributionGiftPool giftService.Update error=", err)
					}
				}
			}
		}
		if totalNum > 0 {
			// 预加载缓存数据
			giftService.GetAll(true)
		}
	}
	return totalNum
}

// 发奖，指定的奖品是否还可以发出来奖品
func PrizeGift(id, leftNum int) bool {
	ok := false
	//缓存减一可以发奖
	ok = prizeServGift(id)
	if ok {
		// 更新数据库，减少奖品的库存
		giftService := services.NewGiftService()
		rows, err := giftService.DecrLeftNum(id, 1)
		//小于1就是没有更新
		if rows < 1 || err != nil {
			log.Println("prizedata.PrizeGift giftService.DecrLeftNum error=", err, ", rows=", rows)
			// 数据更新失败，不能发奖
			return false
		}
	}
	return ok

}

// 获取当前奖品池中的奖品数量
func GetGiftPoolNum(id int) int {
	num := 0
	num = getServGiftPoolNum(id)
	return num
}

// 优惠券类的发放，从redis
func PrizeCodeDiff(id int, codeService services.CodeService) string {
	return prizeServCodeDiff(id, codeService)
	//分布式锁实现线程安全
	//lockUid := 0-id-100000000
	//LockLucky(lockUid)
	//defer UnlockLucky(lockUid)
	////不太好拿到上一次的编码，所以这里先设置为0
	//codeId := 0
	//codeInfo := codeService.NextUsingCode(id,codeId)
	//if codeInfo != nil && codeInfo.Id>0{
	//	//找到了更新状态为2表示发放出去了
	//	codeInfo.SysStatus=2
	//	codeInfo.SysUpdated=comm.NowUnix()
	//	codeService.Update(codeInfo,nil)
	//
	//}else{
	//	log.Printf("prizedata.prizeServCodeDiff num codeInfo,gift_id=",id)
	//}
	//return codeInfo.Code
}

// 优惠券发放，使用redis的方式发放

// 获取当前的缓存中编码数量
// 返回，剩余编码数量，缓冲中编码数量
func GetCacheCodeNum(id int, codeService services.CodeService) (int, int) {
	// 数据库
	num := 0
	//缓存
	cacheNum := 0
	// 统计数据库中有效编码数量
	list := codeService.Search(id)
	if len(list) > 0 {
		for _, data := range list {
			if data.SysStatus == 0 {
				num++
			}
		}
	}

	// redis中缓存的key值
	key := fmt.Sprintf("gift_code_%d", id)
	cacheObj := datasource.InstanceCache()
	//SCARD统计
	rs, err := cacheObj.Do("SCARD", key)
	if err != nil {
		log.Println("prizedata.RecacheCodes RENAME error=", err)
	} else {
		cacheNum = int(comm.GetInt64(rs, 0))
	}

	return num, cacheNum
}

// 导入新的优惠券编码
func ImportCacheCodes(id int, code string) bool {
	// 集群版本需要放入到redis中
	// [暂时]本机版本的就直接从数据库中处理吧
	// redis中缓存的key值
	key := fmt.Sprintf("gift_code_%d", id)
	cacheObj := datasource.InstanceCache()
	_, err := cacheObj.Do("SADD", key, code)
	if err != nil {
		log.Println("prizedata.RecacheCodes SADD error=", err)
		return false
	} else {
		return true
	}
}

// 重新整理优惠券的编码到缓存中
func RecacheCodes(id int, codeService services.CodeService) (sucNum, errNum int) {
	// 集群版本需要放入到redis中
	// [暂时]本机版本的就直接从数据库中处理吧
	list := codeService.Search(id)
	//得到这个奖品所有劵
	if list == nil || len(list) <= 0 {
		return 0, 0
	}
	// redis中缓存的key值
	key := fmt.Sprintf("gift_code_%d", id)
	cacheObj := datasource.InstanceCache()
	//key不是主键。是gift_id
	//临时缓存再正式缓存覆盖，比如数据酷手动删除十个编码，这十个编码在正式key里存在的，但是删了的话从数据库查出来就会少掉十个编码，SADD直接在正式操作上操作这个十个编码不会被删除，在临时key操作就不会把之前的十个编码存进去
	//临时缓存
	tmpKey := "tmp_" + key
	for _, data := range list {
		if data.SysStatus == 0 {
			//不同编码
			code := data.Code
			_, err := cacheObj.Do("SADD", tmpKey, code)
			if err != nil {
				log.Println("prizedata.RecacheCodes SADD error=", err)
				errNum++
			} else {
				sucNum++
			}
		}
	}
	//正式缓存，主键id，RENAME用于修改key名称
	_, err := cacheObj.Do("RENAME", tmpKey, key)
	if err != nil {
		log.Println("prizedata.RecacheCodes RENAME error=", err)
	}
	return sucNum, errNum
}

// 将给定的奖品数量分布到这一天的时间内
//一天的发奖计划
// 结构为： [hour][minute]num
func getGiftPrizeDataOneDay(num int) map[int][60]int {
	rs := make(map[int][60]int)
	//计算24小时
	hourData := [24]int{}
	// 分别将奖品分布到24个小时内
	if num > 100 {
		// 奖品数量多的时候，直接按照百分比计算出来
		for _, h := range conf.PrizeDataRandomDayTime {
			//每个小时的比分
			hourData[h]++
		}
		for h := 0; h < 24; h++ {
			d := hourData[h]
			n := num * d / 100 //也就是总数除以100(概率比分)在乘以这个小时所占的发奖比分，就是这个小时的发就爱那个计划数量
			hourData[h] = n
			num -= n
		}
	}
	// 奖品数量少的时候，或者剩下了一些没有分配，需要用到随即概率来计算
	//上面分配后剩余的
	for num > 0 {
		num--
		// 通过随机数确定奖品落在哪个小时
		hourIndex := comm.Random(100)
		h := conf.PrizeDataRandomDayTime[hourIndex]
		hourData[h]++
	}
	// 将每个小时内的奖品数量分配到60分钟
	for h, hnum := range hourData {
		//数量
		if hnum <= 0 {
			continue
		}
		//每分钟/数量
		minuteData := [60]int{}
		if hnum >= 60 {
			avgMinute := hnum / 60
			for i := 0; i < 60; i++ {
				minuteData[i] = avgMinute
			}
			hnum -= avgMinute * 60
		}
		//平均分配一个小时后剩余的
		// 剩下的数量不多的时候，随机到各分钟内
		for hnum > 0 {
			hnum--
			m := comm.Random(60)
			minuteData[m]++
		}
		rs[h] = minuteData
	}
	return rs
}

// 将每天、每小时、每分钟的奖品数量，格式化成具体到一个时间（分钟）的奖品数量
// 结构为： [day][hour][minute]num：也就是一个大的map,key为第几天，value为一个map，这个map的key为第几小时，value为数量（int类型的长度为60的数组），
//结果为[时间][num](类型为int)
func formatGiftPrizeData(nowTime, dayNum int, prizeData map[int]map[int][60]int) [][2]int {
	rs := make([][2]int, 0)
	//当前时间对应的几点钟
	nowHour := time.Now().Hour()
	// 处理周期内每一天的计划
	for dn := 0; dn < dayNum; dn++ {
		dayData, ok := prizeData[dn]
		if !ok {
			continue
		}
		//第几天的开始时间（第几天当前时间）
		dayTime := nowTime + dn*86400
		// 处理周期每天内，每小时的计划
		for hn := 0; hn < 24; hn++ {
			//循环24小时是从当前时间开始的而不是从0点
			hourData, ok := dayData[(hn+nowHour)%24]
			if !ok {
				continue
			}
			//第几天开始时间的第几个小时的第几分钟（用分钟算）
			hourTime := dayTime + hn*3600
			// 处理周期内，每分钟的计划
			for mn := 0; mn < 60; mn++ {
				//hourData是数组
				num := hourData[mn]
				if num <= 0 {
					continue
				}
				// 找到特定一个时间的计划数据
				//精确到第几秒
				minuteTime := hourTime + mn*60
				//rs是二维数组（也就是多个长度为一的一维数组），每次追加一个长度长度为2的一维数组
				rs = append(rs, [2]int{minuteTime, num})
			}
		}
	}
	return rs
}

// 重置集群的奖品池
func resetServGiftPool() {
	key := "gift_pool"
	cacheObj := datasource.InstanceCache()
	_, err := cacheObj.Do("DEL", key)
	if err != nil {
		log.Println("prizedata.resetServGiftPool DEL error=", err)
	}
}

// 根据计划数据，往奖品池增加奖品数量，并发处理
func incrGiftPool(id, num int) int {
	return incrServGiftPool(id, num)
}

// 往奖品池增加奖品数量，redis缓存，根据计划数据
func incrServGiftPool(id, num int) int {
	key := "gift_pool"
	cacheObj := datasource.InstanceCache()
	rtNum, err := redis.Int64(cacheObj.Do("HINCRBY", key, id, num))
	if err != nil {
		log.Println("prizedata.incrServGiftPool error=", err)
		return 0
	}
	// 保证加入的库存数量正确的被加入到池中
	if int(rtNum) < num {
		// 加少了，补偿一次
		num2 := num - int(rtNum)
		rtNum, err = redis.Int64(cacheObj.Do("HINCRBY", key, id, num2))
		if err != nil {
			log.Println("prizedata.incrServGiftPool2 error=", err)
			return 0
		}
	}
	return int(rtNum)
}

// 发奖，redis缓存,还有没有的发
func prizeServGift(id int) bool {
	key := "gift_pool"
	cacheObj := datasource.InstanceCache()
	rs, err := cacheObj.Do("HINCRBY", key, id, -1)
	if err != nil {
		log.Println("prizedata.prizeServGift error=", err)
		return false
	}
	num := comm.GetInt64(rs, -1)
	if num >= 0 {
		return true
	} else {
		return false
	}
}
func prizeServCodeDiff(id int, codeService services.CodeService) string {
	key := fmt.Sprintf("gift_code_%d", id)
	cacheObj := datasource.InstanceCache()
	//弹出来一个
	rs, err := cacheObj.Do("SPOP", key)
	if err != nil {
		log.Println("prizedata.prizeServCodeDiff error=", err)
		return ""
	}
	code := comm.GetString(rs, "")
	if code == "" {
		log.Printf("prizedata.prizeServCodeDiff rs=%s", rs)
		return ""
	}
	// 更新数据库中的发放状态
	codeService.UpdateByCode(&models.LtCode{
		Code:       code,
		SysStatus:  2,
		SysUpdated: comm.NowUnix(),
	}, nil)
	return code
}

// 设置奖品池的数量
func setGiftPool(id, num int) {
	setServGiftPool(id, num)
}

// 设置奖品池的数量，redis缓存
func setServGiftPool(id, num int) {
	key := "gift_pool"
	cacheObj := datasource.InstanceCache()
	_, err := cacheObj.Do("HSET", key, id, num)
	if err != nil {
		log.Println("prizedata.setServGiftPool error=", err)
	}
}

// 清空奖品的发放计划
func clearGiftPrizeData(giftInfo *models.LtGift, giftService services.GiftService) {
	info := &models.LtGift{
		Id:        giftInfo.Id,
		PrizeData: "",
	}
	err := giftService.Update(info, []string{"prize_data"})
	if err != nil {
		log.Println("prizedata.clearGiftPrizeData giftService.Update",
			info, ", error=", err)
	}
	setGiftPool(giftInfo.Id, 0)
}

// 获取当前奖品池中的奖品数量，从redis中
func getServGiftPoolNum(id int) int {
	key := "gift_pool"
	cacheObj := datasource.InstanceCache()
	rs, err := cacheObj.Do("HGET", key, id)
	fmt.Println("huhuhuhuoo")
	fmt.Println(rs)
	if err != nil {
		log.Println("prizedata.getServGiftPoolNum error=", err)
		return 0
	}
	num := comm.GetInt64(rs, 0)
	return int(num)
}
