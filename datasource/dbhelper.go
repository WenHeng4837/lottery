package datasource

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
	"lottery/conf"
	"sync"
)

//数据库引擎
var masterInstance *xorm.Engine

//并发操作，互斥锁
var dbLock sync.Mutex

// 得到唯一的主库实例,单例模式
func InstanceDbMaster() *xorm.Engine {
	if masterInstance != nil {
		return masterInstance
	}
	dbLock.Lock()
	defer dbLock.Unlock()
	//这里还要判断，因为当第一个锁定后生成后，如果在之前判断锁定还有排队的，这些排队的已经跳过了第一个判断，以防后面这些继续生成所有再判断一次
	if masterInstance != nil {
		return masterInstance
	}
	return NewDbMaster()
}

//创建相关数据库数据库实例对象
func NewDbMaster() *xorm.Engine {
	sourcename := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		conf.DbMaster.User,
		conf.DbMaster.Pwd,
		conf.DbMaster.Host,
		conf.DbMaster.Port,
		conf.DbMaster.Database)
	xorm.NewEngine(conf.DriverName, sourcename)

	instance, err := xorm.NewEngine(conf.DriverName, sourcename)
	if err != nil {
		log.Fatal("dbhelper.InstanceDbMaster NewEngine error ", err)
		return nil
	}
	//用于调试展示SQL语句的相关信息
	//instance.ShowSQL(true)
	instance.ShowSQL(false)
	masterInstance = instance
	return masterInstance
}
