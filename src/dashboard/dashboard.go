package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellojqk/http-proxy-analysis/src/entity"
	"github.com/hellojqk/http-proxy-analysis/src/model"
	"github.com/hellojqk/http-proxy-analysis/src/repository"
	"github.com/hellojqk/http-proxy-analysis/src/service"
	"github.com/spf13/viper"
)

var cli = http.Client{}

// Run .
func Run() {
	repository.InitConn()

	go func() {
		for {
			service.Analysis()
			time.Sleep(time.Second * 10)
		}
	}()
	go func() {
		//每隔24小时清理90天前的对比数据
		service.DeleteProxyLogBeforeCount(5000)
		service.DeleteProxyLogBefore(time.Now().AddDate(0, -3, 0))
		time.Sleep(24 * time.Hour)
	}()

	g := gin.Default()
	group := g.Group("/api")
	g.Static("/assets", "./assets")
	g.Static("/ui", "./ui/dist")
	g.LoadHTMLGlob("./assets/templates/*")
	//查看详情
	g.GET("/proxylog/:id", func(c *gin.Context) {
		proxyLog := &entity.ProxyLog{}
		if err := c.ShouldBindUri(proxyLog); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		err := service.GetProxyLog(proxyLog)
		if err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		fmt.Printf("%v\n", proxyLog)
		c.HTML(http.StatusOK, "detail.tmpl", gin.H{
			"proxyLog": proxyLog,
		})
	})

	//获取代理日志分页数据
	group.GET("/proxylog", func(c *gin.Context) {
		param := &model.ProxyLogListRequestParam{}
		c.BindQuery(param)
		if param.PageSize <= 0 {
			param.PageSize = 30
		}
		if param.Current <= 0 {
			param.Current = 1
		}
		list, total, err := service.ListProxyLog(param)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"total": total, "data": list})
	})

	//重新请求
	group.POST("/proxylog/:id/retry", func(c *gin.Context) {
		proxyLog := &entity.ProxyLog{}
		if err := c.ShouldBindUri(proxyLog); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		err := service.GetProxyLog(proxyLog)
		if err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		var body io.Reader
		if proxyLog.ProxyRequestMethod != "GET" {
			body = strings.NewReader(proxyLog.ProxyRequestBody)
		}
		req, err := http.NewRequest(proxyLog.ProxyRequestMethod, proxyLog.Application.Host+proxyLog.ProxyRequestURL, body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		header := http.Header{}
		json.Unmarshal([]byte(proxyLog.ProxyRequestHeader), &header)
		if header != nil {
			req.Header = header
		}

		_, err = cli.Do(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	//获取应用程序数据
	group.GET("/application", func(c *gin.Context) {
		list, err := service.ListAPP(true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	})

	//获取应用程序数据
	group.GET("/diff_strategy", func(c *gin.Context) {
		list, err := service.ListDiffStrategy()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	})

	//获取应用程序数据
	group.POST("/diff_strategy", func(c *gin.Context) {
		model := &entity.DiffStrategy{}
		if err := c.ShouldBind(model); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		service.InsertDiffStrategy(model)
		c.JSON(http.StatusOK, gin.H{})
	})

	//获取应用程序数据
	group.DELETE("/diff_strategy/:id", func(c *gin.Context) {
		model := &entity.DiffStrategy{}
		if err := c.ShouldBindUri(model); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		service.DeleteDiffStrategy(model)
		c.JSON(http.StatusOK, gin.H{})
	})

	//获取应用程序API数据
	group.GET("/application/:id/api", func(c *gin.Context) {
		app := &entity.Application{}
		if err := c.ShouldBindUri(app); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		list, err := service.ListAPI(app.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	})

	// api字段部分更新
	group.PATCH("/api/:id", func(c *gin.Context) {
		model := &entity.API{}
		if err := c.ShouldBindUri(model); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		var modelMap = make(map[string]interface{})
		if err := c.ShouldBindJSON(&modelMap); err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		err := service.UpdateAPI(model.ID, modelMap)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	group.POST("/login/account", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":           "ok",
			"type":             "account",
			"currentAuthority": "admin",
		})
	})
	group.GET("/currentUser", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":   "Admin",
			"avatar": "https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png",
			"userid": "00000001",
			"email":  "wy27520@gmail.com",
		})
	})

	g.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui")
		c.Abort()
	})

	log.Println(viper.GetInt("dashboardPort"))
	g.Run(fmt.Sprintf(":%d", viper.GetInt("dashboardPort")))
}
