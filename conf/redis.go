package conf

//redis的配置，跟配置mysql差不多
type RdsConfig struct {
	Host      string
	Port      int
	User      string
	Pwd       string
	IsRunning bool
}

var RdsCacheList = []RdsConfig{
	{
		Host: "127.0.0.1",
		Port: 6379,
		//没有用户名
		User:      "",
		Pwd:       "120120",
		IsRunning: true,
	},
}
var RdsCache RdsConfig = RdsCacheList[0]
