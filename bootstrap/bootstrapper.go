package bootstrap

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"
	"io/ioutil"
	"lottery/conf"
	"lottery/cron"
	"time"
)

// 定义一个配置器,类型是func
type Configurator func(bootstrapper *Bootstrapper)

// 使用Go内建的嵌入机制(匿名嵌入)，允许类型之前共享代码和数据
// （Bootstrapper继承和共享 iris.Application ）
type Bootstrapper struct {
	//这里是内置继承
	*iris.Application
	AppName      string
	AppOwner     string
	AppSpawnDate time.Time

	Sessions *sessions.Sessions
}

//实例化
// New returns a new Bootstrapper.
func New(appName, appOwner string, cfgs ...Configurator) *Bootstrapper {
	b := &Bootstrapper{
		AppName:      appName,
		AppOwner:     appOwner,
		AppSpawnDate: time.Now(),
		Application:  iris.New(),
	}
	//遍历cfgs这个切片，Configurator类型是func，直接调用传参
	for _, cfg := range cfgs {
		cfg(b)
	}

	return b
}

// SetupViews loads the templates.模板初始化
func (b *Bootstrapper) SetupViews(viewsDir string) {
	//模板引擎，目录，扩展名，文件
	htmlEngine := iris.HTML(viewsDir, ".html").Layout("shared/layout.html")
	// 每次重新加载模版（线上关闭它）
	htmlEngine.Reload(true)
	// 给模版内置各种定制的方法
	htmlEngine.AddFunc("FromUnixtimeShort", func(t int) string {
		dt := time.Unix(int64(t), int64(0))
		return dt.Format(conf.SysTimeformShort)
	})
	htmlEngine.AddFunc("FromUnixtime", func(t int) string {
		dt := time.Unix(int64(t), int64(0))
		return dt.Format(conf.SysTimeform)
	})
	// 把模板引擎放进去
	b.RegisterView(htmlEngine)
}

// SetupSessions initializes the sessions, optionally.
//func (b *Bootstrapper) SetupSessions(expires time.Duration, cookieHashKey, cookieBlockKey []byte) {
//	b.Sessions = sessions.New(sessions.Config{
//		Cookie:   "SECRET_SESS_COOKIE_" + b.AppName,
//		Expires:  expires,
//		Encoding: securecookie.New(cookieHashKey, cookieBlockKey),
//	})
//}

//// SetupWebsockets prepares the websocket server.
//func (b *Bootstrapper) SetupWebsockets(endpoint string, onConnection websocket.ConnectionFunc) {
//	ws := websocket.New(websocket.Config{})
//	ws.OnConnection(onConnection)
//
//	b.Get(endpoint, ws.Handler())
//	b.Any("/iris-ws.js", func(ctx iris.Context) {
//		ctx.Write(websocket.ClientSource)
//	})
//}

//异常处理
// SetupErrorHandlers prepares the http error handlers
// `(context.StatusCodeNotSuccessful`,  which defaults to < 200 || >= 400 but you can change it).
func (b *Bootstrapper) SetupErrorHandlers() {
	b.OnAnyErrorCode(func(ctx iris.Context) {
		err := iris.Map{
			"app":     b.AppName,
			"status":  ctx.GetStatusCode(),
			"message": ctx.Values().GetString("message"),
		}
		//判断输出方式，如果参数包含json就用json输出错误信息
		if jsonOutput := ctx.URLParamExists("json"); jsonOutput {
			ctx.JSON(err)
			return
		}
		//否则用模板格式输出
		ctx.ViewData("Err", err)
		ctx.ViewData("Title", "Error")
		ctx.View("shared/error.html")
	})
}

// Configure accepts configurations and runs them inside the Bootstraper's context.
//配置方法
func (b *Bootstrapper) Configure(cs ...Configurator) {
	for _, c := range cs {
		c(b)
	}
}

// 启动计划任务服务
func (b *Bootstrapper) setupCron() {

	if conf.RunningCrontabService {
		cron.ConfigueAppOneCron()
	} else {
		cron.ConfigueAppAllCron()
	}

}

//定义两个常量
const (

	// StaticAssets is the root directory for public assets like images, css, js.(StaticAssets是公共资源根目录)
	StaticAssets = "./public"
	// Favicon is the relative 9to the "StaticAssets") favicon path for our app.（Favicon是相对于“StaticAssets”)的应用程序的Favicon路径。）
	Favicon = "/favicon.ico"
)

// Bootstrap prepares our application.
//
// Returns itself.
func (b *Bootstrapper) Bootstrap() *Bootstrapper {
	b.SetupViews("./views")
	//b.SetupSessions(24*time.Hour,
	//	[]byte("the-big-and-secret-fash-key-here"),
	//	[]byte("lot-secret-of-characters-big-too"),
	//)
	b.SetupErrorHandlers()

	// static files
	b.Favicon(StaticAssets + Favicon)
	//视频里的b.StaticWeb过期了
	b.HandleDir(StaticAssets[1:], StaticAssets)
	indexHtml, err := ioutil.ReadFile(StaticAssets + "/index.html")
	if err == nil {
		b.StaticContent(StaticAssets[1:]+"/", "text/html",
			indexHtml)
	}
	// 不要把目录末尾"/"省略掉
	iris.WithoutPathCorrectionRedirection(b.Application)

	// crontab,启动计划任务
	b.setupCron()

	// middleware, after static files
	b.Use(recover.New())
	b.Use(logger.New())

	return b
}

// Listen starts the http server with the specified "addr".
func (b *Bootstrapper) Listen(addr string, cfgs ...iris.Configurator) {
	//监听什么地址，配置是什么
	b.Run(iris.Addr(addr), cfgs...)
}
