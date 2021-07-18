package models

// 站点中与浏览器交互的用户模型
//用于登录的
type ObjLoginuser struct {
	Uid      int
	Username string
	//时间
	Now int
	Ip  string
	//签名
	Sign string
}
