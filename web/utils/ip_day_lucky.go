/**
 * 同一个IP抽奖，每天的操作限制，本地或者redis缓存
 */
package utils

import (
	"fmt"
	"log"
	"math"
	"time"

	"lottery/comm"
	"lottery/datasource"
)

const ipFrameSize = 2

//验证IP今日的参与次数
func init() {
	//// IP当天的统计数，整点归零，设置定时器
	duration := comm.NextDayDuration()
	time.AfterFunc(duration, resetGroupIpList)
	//本地开发测试的时候，每次启动归零
	resetGroupIpList()
}

// 重置单机IP今天次数,每日凌晨计数器归零
func resetGroupIpList() {
	log.Println("ip_day_lucky.resetGroupIpList start")
	cacheObj := datasource.InstanceCache()
	for i := 0; i < ipFrameSize; i++ {
		key := fmt.Sprintf("day_ips_%d", i)
		cacheObj.Do("DEL", key)
	}
	log.Println("ip_day_lucky.resetGroupIpList stop")
	// IP当天的统计数，整点归零，设置定时器
	duration := comm.NextDayDuration()
	time.AfterFunc(duration, resetGroupIpList)
}

// 今天的IP抽奖次数递增，返回递增后的数值
func IncrIpLucyNum(strIp string) int64 {
	ip := comm.Ip4toInt(strIp)
	//把ip地址做散列处理
	i := ip % ipFrameSize
	// 集群的redis统计数递增
	return incrServIpLucyNum(i, ip)
}

//
func incrServIpLucyNum(i, ip int64) int64 {
	key := fmt.Sprintf("day_ips_%d", i)
	cacheObj := datasource.InstanceCache()
	rs, err := cacheObj.Do("HINCRBY", key, ip, 1)
	if err != nil {
		log.Println("ip_day_lucky redis HINCRBY err=", err)
		return math.MaxInt32
	} else {
		return rs.(int64)
	}
}

//今天的IP抽奖次数递增，返回递增后的数值
func IncrIpLuckyNum(strIp string) int64 {
	ip := comm.Ip4toInt(strIp)
	//把ip地址做散列处理
	i := ip % ipFrameSize
	//复用缓存key值
	key := fmt.Sprintf("day_ips_%d", i)
	cacheObj := datasource.InstanceCache()
	rs, err := cacheObj.Do("HINCRBY", key, ip, 1)
	if err != nil {
		log.Println("ip_day_lucky redis HINCRBY error=", err)
		return math.MaxInt32
	} else {
		return rs.(int64)
	}
}
