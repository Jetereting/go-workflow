package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	router "go-workflow/workflow-router"

	config "go-workflow/workflow-config"

	model "go-workflow/workflow-engine/model"
)

// 配置
var conf = *config.Config

func main() {
	mux := router.Mux
	// 启动数据库连接
	model.Setup()
	// 启动定时任务
	//service.CronJobs()
	// 启动服务
	readTimeout, err := strconv.Atoi(conf.ReadTimeout)
	if err != nil {
		panic(err)
	}
	writeTimeout, err := strconv.Atoi(conf.WriteTimeout)
	if err != nil {
		panic(err)
	}
	server := &http.Server{
		Addr:           fmt.Sprintf("%s", conf.Port),
		Handler:        mux,
		ReadTimeout:    time.Duration(readTimeout * int(time.Second)),
		WriteTimeout:   time.Duration(writeTimeout * int(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("the application start up at %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}

}
