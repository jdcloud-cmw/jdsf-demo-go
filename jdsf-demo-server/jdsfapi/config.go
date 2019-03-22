package jdsfapi

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"strings"
)

var(
	JDSFGlobalConfig *JDSFConfig
)




type JDSFConfig struct {
	Consul *ConsulConfig  `yaml:"consul"`
	Tracing *TraceConfig  `yaml:"trace"`
	AppConfig *AppConfig  `yaml:"app"`
}

func NewJDSFConfig() *JDSFConfig {
	consulConfig := new(ConsulConfig)
	consulConfig.Scheme="http"
	consulConfig.Port=8500
	consulDiscover:= new (ConsulDiscover)
	consulDiscover.Enable=false
	consulDiscover.CheckUrl = "/api/health/check"
	consulConfig.Discover = consulDiscover
	traceConfig :=new(TraceConfig)
	traceConfig.Enable = false
	jdsfConfig :=new(JDSFConfig)
	jdsfConfig.Consul = consulConfig
	jdsfConfig.Tracing = traceConfig
	return jdsfConfig
}

type ConsulConfig struct {
	Scheme string `yaml:"scheme"`
	Address string `yaml:"address"`
	Port    int32 `yaml:"port"`
	Discover *ConsulDiscover `yaml:"discover"`
}

type ConsulDiscover struct {
	Enable bool `yaml:"enable"`
	ServiceInstanceId string `yaml:"instanceId"`
	CheckUrl string `yaml:"checkUrl"`
}


type TraceConfig struct {
	Enable bool `yaml:"enable"`
	SimpleType string `yaml:"simpleType"`
	SimpleRate  float64 `yaml:"simpleRate"`
	TraceUdpAddress string `yaml:"traceUdpAddress"`
	TraceUdpPort int `yaml:"traceUdpPort"`
	TraceHttpAddress string `yaml:"traceHttpAddress"`
	TraceHttpPort int `yaml:"traceHttpPort"`
}

type AppConfig struct {
	AppName string `yaml:"appName"`
	HostIp string `yaml:"hostIp"`
	ServerPort int32 `yaml:"serverPort"`
}



func (j *JDSFConfig)LoadConfig(configFilePath string) *JDSFConfig  {
	appConfig:=new(JDSFConfig)
	configFile,err := ioutil.ReadFile(configFilePath)

	if err !=nil{
		fmt.Print(err)
		return nil
	}
	yamlerr:=yaml.Unmarshal(configFile,appConfig)
	if yamlerr !=nil{
		fmt.Print(yamlerr)
		return nil
	}
	j.AppConfig = appConfig.AppConfig
	if appConfig.Tracing.Enable{
		j.Tracing = appConfig.Tracing
	}
	if appConfig.Consul.Address != ""{
		j.Consul.Address = appConfig.Consul.Address
		if appConfig.Consul.Port>0{
			j.Consul.Port = appConfig.Consul.Port
		}
		if appConfig.Consul.Scheme !=""{
			j.Consul.Scheme = appConfig.Consul.Scheme
		}
		if appConfig.Consul.Discover != nil && appConfig.Consul.Discover.Enable{
			j.Consul.Discover.Enable=true
			if appConfig.Consul.Discover.CheckUrl != ""{
				j.Consul.Discover.CheckUrl = appConfig.Consul.Discover.CheckUrl
			}
			if appConfig.Consul.Discover.ServiceInstanceId !=""{
				j.Consul.Discover.ServiceInstanceId = appConfig.Consul.Discover.ServiceInstanceId
			}else{
				if j.AppConfig.AppName!="" && j.AppConfig.ServerPort>0{
					 instanceUUID := strings.Replace(uuid.New().String(),"-","",-1)
					j.Consul.Discover.ServiceInstanceId = j.AppConfig.AppName+"-"+strconv.Itoa(int(j.AppConfig.ServerPort))+instanceUUID
				}
			}
		}
	}
	JDSFGlobalConfig = j
	return j
}