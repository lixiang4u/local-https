package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/local-https/helper"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if !helper.WindowsTokenElevated() {
		helper.ExitMsg("请使用管理员权限打开")
	}

	certName, proxyMap, err := initConfig()
	if err != nil {
		return
	}
	if err != nil {
		helper.ExitMsg(fmt.Sprintf("配置文件异常：%s", err.Error()))
		return
	}

	cert, key, err := helper.MakeDomainCertificate(certName, helper.MapKeys(proxyMap))
	if err != nil {
		helper.ExitMsg(fmt.Sprintf("证书生成失败：%s", err.Error()))
		return
	}
	_, err = helper.ReplaceCertToRoot(cert)
	if err != nil {
		helper.ExitMsg(fmt.Sprintf("导入证书失败：%s", err.Error()))
		return
	}
	//log.Println("[导入证书信息]", string(output))

	var p = 8060
	for tmpHost, tmpRedirect := range proxyMap {
		tmpHost = strings.TrimSpace(tmpHost)
		if !helper.CheckHost(tmpHost) {
			continue
		}
		if helper.ParseHost(tmpRedirect) != "" {
			// 使用配置的转发地址
			proxyMap[tmpHost] = tmpRedirect
			log.Println(fmt.Sprintf("[转发地址] %s -> %s", tmpHost, tmpRedirect))
		} else {
			// 启动虚拟web服务
			p = helper.NextUsefulPort(p)
			go runLocalHttpServer(p, tmpHost)
			// 使用默认的转发地址
			proxyMap[tmpHost] = fmt.Sprintf("http://127.0.0.1:%d", p)
			log.Println(fmt.Sprintf("[转发地址] %s -> 127.0.0.1:%d", tmpHost, p))
		}
		time.Sleep(time.Second / 3)
		_ = helper.UpdateWindowsHosts(fmt.Sprintf("127.0.0.1	%s", tmpHost))
	}

	runReverseProxyServer(proxyMap, cert, key)

	helper.ExitWithSigExit()
}

func runLocalHttpServer(port int, domain string) {
	_ = helper.MkdirAll(filepath.Join(helper.AppPath(), "www", domain, "1.txt"))
	gin.SetMode(gin.ReleaseMode)
	var app = gin.Default()
	app.Use(gin.Recovery())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Hit-Cache"},
		AllowCredentials: true,
	}))
	app.StaticFS("/", gin.Dir(filepath.Join(helper.AppPath(), "www", domain), true))
	app.NoRoute(func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("[http://127.0.0.1:%d] %s", port, time.Now().String()))
	})
	_ = helper.WriteFileContent(
		filepath.Join(helper.AppPath(), "www", domain, "index.html"),
		[]byte(fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="UTF-8"><title>%s</title></head><body><h1>已经转发 %s -> http://127.0.0.1:%d</h1></body></html>`, domain, domain, port)),
	)
	if err := app.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Println(fmt.Sprintf("[http://127.0.0.1:%d] 启动失败 %s", port, err.Error()))
	}
}

func runWebServer(domain, cert, key string) {
	var app = gin.Default()
	app.Use(gin.Recovery())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Hit-Cache"},
		AllowCredentials: true,
	}))
	app.NoRoute(func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("[%s] %s", domain, time.Now().String()))
	})

	if err := app.RunTLS(fmt.Sprintf(":443"), cert, key); err != nil {
		log.Println(fmt.Sprintf("站点(%s)启动失败：%s", domain, err.Error()))
	}
}

func runReverseProxyServer(proxyHostMap map[string]string, cert, key string) {
	var proxyHandlers = helper.NewHostReverseProxyHandlerMap(proxyHostMap)

	//http.HandleFunc("/api/cctv", func(w http.ResponseWriter, r *http.Request) {
	//	log.Println("[自定义处理器]", r.Host, r.RequestURI)
	//})

	//var registerHandler = make([])
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 查找对应的转发处理器
		if handler, ok := proxyHandlers[r.Host]; ok {
			log.Println("[local-https-req]", fmt.Sprintf("http(s)://%s/%s", strings.TrimRight(r.Host, "/"), strings.TrimLeft(r.RequestURI, "/")))
			if handler == nil {
				return
			}
			handler.ServeHTTP(w, r)
		} else {
			// 如果没有找到对应的域名映射
			http.Error(w, fmt.Sprintf("[%s] not found", r.Host), http.StatusNotFound)
		}
	})
	go func() {
		log.Println("[启动本地http服务]")
		log.Fatal(http.ListenAndServe(":80", nil))
	}()
	go func() {
		log.Println("[启动本地https服务]")
		log.Fatal(http.ListenAndServeTLS(":443", cert, key, nil))
	}()
}

func initConfig() (certName string, proxyMap map[string]string, err error) {
	viper.SetConfigFile(filepath.Join(helper.AppPath(), "config.json"))
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	if err = viper.UnmarshalKey("cert_name", &certName); err != nil {
		return
	}
	if err = viper.UnmarshalKey("proxy", &proxyMap); err != nil {
		return
	}
	return
}
