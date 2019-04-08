package jdsfapi

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/jdsfapi/util"
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

func (j *JDSFConfig)GetAppConsulConfigKey() string  {
	appName := JDSFGlobalConfig.AppConfig.AppName
	profileName := JDSFGlobalConfig.Consul.Config.Profile
	return  getAppConsulConfigKey(appName,profileName)
}

func getAppConsulConfigKey(appName string,profileName string) string {
	if appName == ""{
		appName = "application"
	}
	if profileName != ""{
		profileName = ","+profileName
	}
	consulConfigKey :="config/"+appName+profileName+"/data"
	return consulConfigKey
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
	Port    int32 `yaml:"port"`
	Discover *ConsulDiscover `yaml:"discover"`
	Address string `yaml:"address"`
	Config *ConsulConfigServer `yaml:"config"`
}

type ConsulDiscover struct {
	Enable bool `yaml:"enable"`					  //是否启用服务发现
	ServiceInstanceId string `yaml:"instanceId"`  //服务实例 id
	CheckUrl string `yaml:"checkUrl"`			//健康检查 url
	InstanceZone string `yaml:"instanceZone"`  // 服务所在的可用区
	ServiceTTLTime int `yaml:"serviceTTLTime"` // 自动更新服务实例缓存列表时间
}

type ConsulConfigServer  struct{
	Enable bool `yaml:"enable"`
	Profile string `yaml:"profile"`
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
	if appConfig.Consul.Address != ""{
		j.Consul.Address = appConfig.Consul.Address
		if appConfig.Consul.Port>0{
			j.Consul.Port = appConfig.Consul.Port
		}
		if appConfig.Consul.Scheme !=""{
			j.Consul.Scheme = appConfig.Consul.Scheme
		}
		if appConfig.Consul.Config.Enable{
			defaultConfig := api.DefaultConfig()
			consulPortStr := strconv.Itoa(int( appConfig.Consul.Port))
			defaultConfig.Address = appConfig.Consul.Address+":"+consulPortStr
			fmt.Println(defaultConfig.Address)
			defaultConfig.Scheme = appConfig.Consul.Scheme
			client, err := api.NewClient(defaultConfig)
			if err != nil{
				fmt.Println(err)
				return nil
			}
			configKey := getAppConsulConfigKey(appConfig.AppConfig.AppName,appConfig.Consul.Config.Profile)
			kvPair,_,err := client.KV().Get(configKey,nil)
			if err == nil{
				fmt.Println(err)
				return nil
			}
			consulAppConfig:=new(JDSFConfig)
			yamlErr :=yaml.Unmarshal(kvPair.Value,consulAppConfig)
			if yamlErr !=nil{
				fmt.Print(yamlErr)
				return nil
			}
			consulAppConfig.AppConfig.AppName = appConfig.AppConfig.AppName
			if consulAppConfig.AppConfig.ServerPort<=0{
				consulAppConfig.AppConfig.ServerPort = appConfig.AppConfig.ServerPort
			}

			appConfig = consulAppConfig
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
	j.AppConfig = appConfig.AppConfig
	if j.AppConfig.HostIp == ""||strings.Trim(j.AppConfig.HostIp," ") == ""{
		if j.Consul.Discover.Enable{
			j.AppConfig.HostIp = util.GetHostIpUsePing(j.Consul.Address+":"+strconv.Itoa(int(j.Consul.Port)))
		}else{
			j.AppConfig.HostIp = util.GetHostIp()
		}

	}
	if appConfig.Tracing.Enable{
		j.Tracing = appConfig.Tracing
	}
	JDSFGlobalConfig = j
	return j
}