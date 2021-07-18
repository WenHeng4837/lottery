package dao

import (
	"github.com/go-xorm/xorm"
	"log"
	"lottery/comm"
	"lottery/models"
)

//定义数据库操作引擎
type GiftDao struct {
	engine *xorm.Engine
}

//获取一个GiftDao对象
func NewGiftDao(engine *xorm.Engine) *GiftDao {
	return &GiftDao{
		engine: engine,
	}
}

//根据传进来的id得到奖品对象
func (d *GiftDao) Get(id int) *models.LtGift {
	//根据new一个对象,这里&用的是指针，所以下面查询之后会回填进去
	data := &models.LtGift{Id: id}
	//去数据库查找
	ok, err := d.engine.Get(data)
	//没有错误就正常返回
	if ok && err == nil {
		return data
	} else {
		//有错误就清空id再返回，不然直接返回一个正常具有id的对象，这个对象其他属性都为空
		data.Id = 0
		return data
	}
}

//查询所有奖品对象,返回一个奖品切片因为多个礼品，相当于集合礼品
func (d *GiftDao) GetAll() []models.LtGift {
	//先初始化一个长度为0的切片
	datalist := make([]models.LtGift, 0)
	//去查询，先按状态排序(这个奖品存在)再按位置序号排序(第一个奖品第二个奖品)
	//这里的Find一定要传指针进去
	err := d.engine.Asc("sys_status").Asc("displayorder").Find(&datalist)
	//有错误的时候
	if err != nil {
		log.Println("gift_dao.GetAll err=", err)
		return datalist
	}
	return datalist
}

//计算数量多少
func (d *GiftDao) CountAll() int64 {
	num, err := d.engine.Count(&models.LtGift{})
	//如果有错误
	if err != nil {
		return 0
	} else {
		return num
	}
}

//根据id删掉一个奖品
func (d *GiftDao) Delete(id int) error {
	//这里都要不做物理删除，所以删除就是将奖品状态更新为1，1状态就是删除
	now := comm.NowUnix()
	data := &models.LtGift{Id: id, SysStatus: 1, SysUpdated: now}
	_, err := d.engine.Id(data.Id).Update(data)
	return err
}

//更新,columns []string用来表示哪个字段一定要更新
func (d *GiftDao) Update(data *models.LtGift, columns []string) error {
	//如果这个对象里面的对象空的就默认不做更新，因为一个结构体有默认值，如果默认为空就不更新也就是更新失败，所以这里用columns强制设置字段，设置之后如果某个字段设置为空也会强制执行
	_, err := d.engine.Id(data.Id).MustCols(columns...).Update(data)
	return err
}

//插入一条数据
func (d *GiftDao) Create(data *models.LtGift) error {
	_, err := d.engine.Insert(data)
	return err
}

// 获取到当前可以获取的奖品列表
// 有奖品限定，状态正常，时间期间内
// gtype倒序， displayorder正序
func (d *GiftDao) GetAllUse() []models.LtGift {
	now := comm.NowUnix()
	datalist := make([]models.LtGift, 0)
	err := d.engine.
		Cols("id", "title", "prize_num", "left_num", "prize_code",
			"prize_time", "img", "displayorder", "gtype", "gdata").
		Desc("gtype").
		Asc("displayorder").
		Where("prize_num>=?", 0).    // 有限定的奖品
		Where("sys_status=?", 0).    // 有效的奖品
		Where("time_begin<=?", now). // 时间期内
		Where("time_end>=?", now).   // 时间期内
		Find(&datalist)
	if err != nil {
		return datalist
	} else {
		return datalist
	}
}

//同下
func (d *GiftDao) IncrLeftNum(id, num int) (int64, error) {
	r, err := d.engine.Id(id).
		Incr("left_num", num).
		//Where("left_num=?", num).
		Update(&models.LtGift{Id: id})
	return r, err
}

//id为奖品id,num为要减掉数量
func (d *GiftDao) DecrLeftNum(id, num int) (int64, error) {
	r, err := d.engine.Id(id).
		Decr("left_num", num).
		Where("left_num>=?", num).
		Update(&models.LtGift{Id: id})
	return r, err
}
