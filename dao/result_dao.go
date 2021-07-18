/**
 * 抽奖系统的数据库操作
 */
package dao

import (
	"github.com/go-xorm/xorm"

	"lottery/models"
)

//中奖记录
type ResultDao struct {
	engine *xorm.Engine
}

func NewResultDao(engine *xorm.Engine) *ResultDao {
	return &ResultDao{
		engine: engine,
	}
}

func (d *ResultDao) Get(id int) *models.LtResult {
	data := &models.LtResult{Id: id}
	ok, err := d.engine.Get(data)
	if ok && err == nil {
		return data
	} else {
		data.Id = 0
		return data
	}
}

func (d *ResultDao) GetAll(page, size int) []models.LtResult {
	offset := (page - 1) * size
	datalist := make([]models.LtResult, 0)
	err := d.engine.
		Desc("id").
		Limit(size, offset).
		Find(&datalist)
	if err != nil {
		return datalist
	} else {
		return datalist
	}
}

func (d *ResultDao) CountAll() int64 {
	num, err := d.engine.
		Count(&models.LtResult{})
	if err != nil {
		return 0
	} else {
		return num
	}
}

//根据奖品id分页查询
func (d *ResultDao) GetNewPrize(size int, giftIds []int) []models.LtResult {
	datalist := make([]models.LtResult, 0)
	err := d.engine.
		In("gift_id", giftIds).
		Desc("id").
		Limit(size).
		Find(&datalist)
	if err != nil {
		return datalist
	} else {
		return datalist
	}
}

//根据奖品id分页查询，size, offset
func (d *ResultDao) SearchByGift(giftId, page, size int) []models.LtResult {
	offset := (page - 1) * size
	datalist := make([]models.LtResult, 0)
	err := d.engine.
		Where("gift_id=?", giftId).
		Desc("id").
		Limit(size, offset).
		Find(&datalist)
	if err != nil {
		return datalist
	} else {
		return datalist
	}
}

//根据用户id分页查询
func (d *ResultDao) SearchByUser(uid, page, size int) []models.LtResult {
	offset := (page - 1) * size
	datalist := make([]models.LtResult, 0)
	err := d.engine.
		Where("uid=?", uid).
		Desc("id").
		Limit(size, offset).
		Find(&datalist)
	if err != nil {
		return datalist
	} else {
		return datalist
	}
}

//根据对应的奖品的id得到对应中奖记录总数
func (d *ResultDao) CountByGift(giftId int) int64 {
	num, err := d.engine.
		Where("gift_id=?", giftId).
		Count(&models.LtResult{})
	if err != nil {
		return 0
	} else {
		return num
	}
}

//一个用户的所有中奖记录条数
func (d *ResultDao) CountByUser(uid int) int64 {
	num, err := d.engine.
		Where("uid=?", uid).
		Count(&models.LtResult{})
	if err != nil {
		return 0
	} else {
		return num
	}
}

func (d *ResultDao) Delete(id int) error {
	data := &models.LtResult{Id: id, SysStatus: 1}
	_, err := d.engine.Id(data.Id).Update(data)
	return err
}

func (d *ResultDao) Update(data *models.LtResult, columns []string) error {
	_, err := d.engine.Id(data.Id).MustCols(columns...).Update(data)
	return err
}

func (d *ResultDao) Create(data *models.LtResult) error {
	_, err := d.engine.Insert(data)
	return err
}
