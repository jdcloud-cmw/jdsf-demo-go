package service

import (
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-server/jdsfapi"
	"github.com/opentracing/opentracing-go"
	"log"
	"net/http"
	"strconv"
)

func StartWebServer()  {
	jdsfConfig := jdsfapi.NewJDSFConfig()
	jdsfConfig.LoadConfig("./conf/appConfig.yaml")
	appConfig :=jdsfConfig.AppConfig
	traceConfig := jdsfConfig.Tracing
	jdsfapi.AppTraceGlobalConfig()
	registryClient := jdsfapi.NewRegistryClient()
	registryClient.GetConsulClient()
	registryClient.RegistryService()
	r := NewRouter()
	http.Handle("/", r)
	println(appConfig.ServerPort)
	portStr := strconv.Itoa(int(appConfig.ServerPort))
	log.Println("Starting HTTP service at " + portStr)
	var err error
	if traceConfig.Enable{
		err = http.ListenAndServe(":" + portStr, jdsfapi.Middleware(opentracing.GlobalTracer(),http.DefaultServeMux))
	}else{
		err = http.ListenAndServe(":" + portStr,nil)
	}

	if err != nil {
		log.Println("An error occured starting HTTP listener at port " + portStr)
		log.Println("Error: " + err.Error())
	}
}