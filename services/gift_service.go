package services

import (
	"encoding/json"
	"log"
	"lottery/comm"
	"lottery/dao"
	"lottery/datasource"
	"lottery/models"
	"strconv"
	"strings"
)

//奖品服务类
//定义一个接口
//GiftService对外公开
type GiftService interface {
	//里面定义一些方法以及返回值参数,在这里这些方法还没有被实现
	GetAll(useCache bool) []models.LtGift
	CountAll() int64
	Get(id int, useCache bool) *models.LtGift
	Delete(id int) error
	Update(data *models.LtGift, columns []string) error
	Create(data *models.LtGift) error
	GetAllUse(useCache bool) []models.ObjGiftPrize
	IncrLeftNum(id, num int) (int64, error)
	DecrLeftNum(id, num int) (int64, error)
}

//定义对象,用于实现上面这个接口，小写私有
type giftService struct {
	//需要用到数据库操作,后面如果用到redis还需要Redis的操作对象，这里先这样
	dao *dao.GiftDao
}

//构造方法返回GiftService这个接口
func NewGiftService() GiftService {
	//giftService实现了接口
	return &giftService{
		dao: dao.NewGiftDao(datasource.InstanceDbMaster()),
	}
}

//接下来实现接口里的那些方法
//作用在giftService对象上的
func (s *giftService) GetAll(useCache bool) []models.LtGift {
	if !useCache {
		// 直接读取数据库的方式
		return s.dao.GetAll()
	}
	// 先读取缓存
	gifts := s.getAllByCache()
	if len(gifts) < 1 {
		// 再读取数据库
		gifts = s.dao.GetAll()
		s.setAllByCache(gifts)
	}
	return gifts
}
func (s *giftService) CountAll() int64 {
	// 直接读取数据库的方式
	//return s.dao.CountAll()

	// 缓存优化之后的读取方式
	gifts := s.GetAll(true)
	return int64(len(gifts))
}
func (s *giftService) Get(id int, useCache bool) *models.LtGift {
	if !useCache {
		// 直接读取数据库的方式
		return s.dao.Get(id)
	}
	// 缓存优化之后的读取方式
	gifts := s.GetAll(true)
	for _, gift := range gifts {
		if gift.Id == id {
			return &gift
		}
	}
	return nil
}
func (s *giftService) Delete(id int) error {
	// 先更新缓存
	data := &models.LtGift{Id: id}
	s.updateByCache(data, nil)
	// 再更新数据库
	return s.dao.Delete(id)
}
func (s *giftService) Update(data *models.LtGift, columns []string) error {
	// 先更新缓存
	s.updateByCache(data, columns)
	// 再更新数据库
	return s.dao.Update(data, columns)
}
func (s *giftService) Create(data *models.LtGift) error {
	// 先更新缓存
	s.updateByCache(data, nil)
	// 再更新数据库
	return s.dao.Create(data)
}

// 获取到当前可以获取的奖品列表
// 有奖品限定，状态正常，时间期间内
// gtype倒序， displayorder正序
//useCache bool
func (s *giftService) GetAllUse(useCache bool) []models.ObjGiftPrize {
	list := make([]models.LtGift, 0)
	if !useCache {
		// 直接读取数据库的方式
		list = s.dao.GetAllUse()
	} else {
		// 缓存优化之后的读取方式
		now := comm.NowUnix()
		gifts := s.GetAll(true)
		for _, gift := range gifts {
			if gift.Id > 0 && gift.SysStatus == 0 &&
				gift.PrizeNum >= 0 &&
				gift.TimeBegin <= now &&
				gift.TimeEnd >= now {
				list = append(list, gift)
			}
		}
	}

	if list != nil {
		gifts := make([]models.ObjGiftPrize, 0)
		for _, gift := range list {
			//数据库的gift实体的字段prizecode就是比如100-700这样的
			codes := strings.Split(gift.PrizeCode, "-")
			if len(codes) == 2 {
				// 设置了获奖编码范围 a-b 才可以进行抽奖
				codeA := codes[0]
				codeB := codes[1]
				a, e1 := strconv.Atoi(codeA)
				b, e2 := strconv.Atoi(codeB)
				if e1 == nil && e2 == nil && b >= a && a >= 0 && b < 10000 {
					data := models.ObjGiftPrize{
						Id:           gift.Id,
						Title:        gift.Title,
						PrizeNum:     gift.PrizeNum,
						LeftNum:      gift.LeftNum,
						PrizeCodeA:   a,
						PrizeCodeB:   b,
						Img:          gift.Img,
						Displayorder: gift.Displayorder,
						Gtype:        gift.Gtype,
						Gdata:        gift.Gdata,
					}
					gifts = append(gifts, data)
				}
			}
		}
		return gifts
	} else {
		//不然返回空数组
		return []models.ObjGiftPrize{}
	}
}

//增加奖品剩余数量，也就是更新库存,递增
func (s *giftService) IncrLeftNum(id, num int) (int64, error) {
	return s.dao.IncrLeftNum(id, num)
}

//减少奖品剩余数量，也就是更新库存，递减
func (s *giftService) DecrLeftNum(id, num int) (int64, error) {
	return s.dao.DecrLeftNum(id, num)
}

// 从缓存中获取全部的奖品
func (s *giftService) getAllByCache() []models.LtGift {
	// 集群模式，redis缓存
	key := "allgift"
	rds := datasource.InstanceCache()
	// 读取缓存
	rs, err := rds.Do("GET", key)
	if err != nil {
		log.Println("gift_service.getAllByCache GET key=", key, ", error=", err)
		return nil
	}
	str := comm.GetString(rs, "")
	if str == "" {
		return nil
	}
	// 将json数据反序列化
	datalist := []map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &datalist)
	if err != nil {
		log.Println("gift_service.getAllByCache json.Unmarshal error=", err)
		return nil
	}
	// 格式转换
	gifts := make([]models.LtGift, len(datalist))
	for i := 0; i < len(datalist); i++ {
		data := datalist[i]
		//拿到id
		id := comm.GetInt64FromMap(data, "Id", 0)
		if id <= 0 {
			gifts[i] = models.LtGift{}
		} else {
			gift := models.LtGift{
				Id:           int(id),
				Title:        comm.GetStringFromMap(data, "Title", ""),
				PrizeNum:     int(comm.GetInt64FromMap(data, "PrizeNum", 0)),
				LeftNum:      int(comm.GetInt64FromMap(data, "LeftNum", 0)),
				PrizeCode:    comm.GetStringFromMap(data, "PrizeCode", ""),
				PrizeTime:    int(comm.GetInt64FromMap(data, "PrizeTime", 0)),
				Img:          comm.GetStringFromMap(data, "Img", ""),
				Displayorder: int(comm.GetInt64FromMap(data, "Displayorder", 0)),
				Gtype:        int(comm.GetInt64FromMap(data, "Gtype", 0)),
				Gdata:        comm.GetStringFromMap(data, "Gdata", ""),
				TimeBegin:    int(comm.GetInt64FromMap(data, "TimeBegin", 0)),
				TimeEnd:      int(comm.GetInt64FromMap(data, "TimeEnd", 0)),
				//发奖计划，没有因为非常大没有做序列化存储，所以也读不出来这里就为空
				//PrizeData:    comm.GetStringFromMap(data, "PrizeData", ""),
				PrizeBegin: int(comm.GetInt64FromMap(data, "PrizeBegin", 0)),
				PrizeEnd:   int(comm.GetInt64FromMap(data, "PrizeEnd", 0)),
				SysStatus:  int(comm.GetInt64FromMap(data, "SysStatus", 0)),
				SysCreated: int(comm.GetInt64FromMap(data, "SysCreated", 0)),
				SysUpdated: int(comm.GetInt64FromMap(data, "SysUpdated", 0)),
				SysIp:      comm.GetStringFromMap(data, "SysIp", ""),
			}
			gifts[i] = gift
		}
	}
	return gifts
}

// 将奖品的数据更新到缓存
func (s *giftService) setAllByCache(gifts []models.LtGift) {
	// 集群模式，redis缓存
	strValue := ""
	if len(gifts) > 0 {
		datalist := make([]map[string]interface{}, len(gifts))
		// 格式转换
		for i := 0; i < len(gifts); i++ {
			gift := gifts[i]
			data := make(map[string]interface{})
			data["Id"] = gift.Id
			data["Title"] = gift.Title
			data["PrizeNum"] = gift.PrizeNum
			data["LeftNum"] = gift.LeftNum
			data["PrizeCode"] = gift.PrizeCode
			data["PrizeTime"] = gift.PrizeTime
			data["Img"] = gift.Img
			data["Displayorder"] = gift.Displayorder
			data["Gtype"] = gift.Gtype
			data["Gdata"] = gift.Gdata
			data["TimeBegin"] = gift.TimeBegin
			data["TimeEnd"] = gift.TimeEnd
			//data["PrizeData"] = gift.PrizeData
			data["PrizeBegin"] = gift.PrizeBegin
			data["PrizeEnd"] = gift.PrizeEnd
			data["SysStatus"] = gift.SysStatus
			data["SysCreated"] = gift.SysCreated
			data["SysUpdated"] = gift.SysUpdated
			data["SysIp"] = gift.SysIp
			datalist[i] = data
		}
		//序列化
		str, err := json.Marshal(datalist)
		if err != nil {
			log.Println()
		}
		strValue = string(str)
	}
	key := "allgift"
	rds := datasource.InstanceCache()
	// 更新缓存
	_, err := rds.Do("SET", key, strValue)
	if err != nil {
		log.Println("gift_service.setAllByCache SET key=", key,
			", value=", strValue, ", error=", err)
	}
}

// 数据更新，需要更新缓存，直接清空缓存数据,删除之后会自动调用设置缓存函数
func (s *giftService) updateByCache(data *models.LtGift, columns []string) {
	if data == nil || data.Id <= 0 {
		return
	}
	// 集群模式，redis缓存
	key := "allgift"
	rds := datasource.InstanceCache()
	// 删除redis中的缓存
	rds.Do("DEL", key)
}
