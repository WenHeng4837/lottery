package conf

//数据库配置
//定义驱动，mysql
const DriverName = "mysql"

//定义一个结构配置数据库
type DbConfig struct {
	Host     string
	Port     int
	User     string
	Pwd      string
	Database string
	//状态，不一定能用得上
	IsRunning bool
}

//配置不一定只有一个所以用切片
var DbMasterList = []DbConfig{
	{
		Host:      "127.0.0.1",
		Port:      3306,
		User:      "root",
		Pwd:       "120120",
		Database:  "lottery",
		IsRunning: true,
	},
}
var DbMaster DbConfig = DbMasterList[0]
