package jdsfapi

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)
const	consulHostEnv = "CONSUL_HOST"
const 	consulPortEnv  ="CONSUL_PORT"


type RegistryClient struct {
	Address string
	Port int
	Scheme string
	Client  *api.Client
	serviceCache *ServiceCacheMap
	cacheTTL bool
	timer *time.Ticker
}


type ServiceCacheMap struct {
	v map[string][]*api.ServiceEntry
	sync.RWMutex
}

func NewServiceCacheMap() * ServiceCacheMap {
	serviceCacheMap := new(ServiceCacheMap)
	serviceCacheMap.v = make(map[string][]*api.ServiceEntry)
	return serviceCacheMap
}

func (sc *ServiceCacheMap)Put(key string,value []*api.ServiceEntry)  {
	sc.Lock()
	defer sc.Unlock()
	sc.v[key]=value
}
func (sc *ServiceCacheMap)Get(key string) []*api.ServiceEntry {
	sc.RLock()
	defer   sc.RUnlock()
	return sc.v[key]
}

func (sc *ServiceCacheMap)Load(key string) ( []*api.ServiceEntry ,bool) {
	sc.RLock()
	defer   sc.RUnlock()
	value,ok := sc.v[key]
	return value,ok
}

func (sc *ServiceCacheMap)Length() int {
	sc.RLock()
	defer   sc.RUnlock()
	return len(sc.v)
}

func (sc *ServiceCacheMap)Clear()  {
	sc.RLock()
	defer  sc.RUnlock()
	sc.v = make(map[string][]*api.ServiceEntry)
}


var(
	JDSFRegistryClient *RegistryClient
)




func NewRegistryClient() *RegistryClient {
	registryClient := new (RegistryClient)
	consulConfig :=JDSFGlobalConfig.Consul
	host :=	os.Getenv(consulHostEnv)
	port := os.Getenv(consulPortEnv)
	if host !=""{
		consulConfig.Address = host
	}


	if port !=""{
		portValue ,err := strconv.ParseInt(port,10,32)
		if err != nil{
			println(err)
		}else{
			if portValue >0{
				consulConfig.Port = int32(portValue)
			}
		}
	}
	registryClient.Port = int(consulConfig.Port)
	registryClient.Address = consulConfig.Address
	ttlTime := JDSFGlobalConfig.Consul.Discover.ServiceTTLTime
	if ttlTime <= 0{
		ttlTime = 30
	}
	if registryClient.serviceCache == nil{
		registryClient.serviceCache = NewServiceCacheMap()
	}
	registryClient.timer = time.NewTicker(time.Second * time.Duration(ttlTime))
	registryClient.serviceInfoTTL()
	JDSFRegistryClient = registryClient
	return registryClient
}

func (r *RegistryClient)serviceInfoTTL()  {
	if !r.cacheTTL {
		r.cacheTTL = true
		go func() {
			for {
				select {
				case <-r.timer.C:
					if r.serviceCache.Length()>0{
						for key  := range r.serviceCache.v {
							serviceEntrys := r.ServiceRegistryCheck(key)
							if serviceEntrys!=nil && len(serviceEntrys) >0{
								r.serviceCache.Put(key,serviceEntrys)
							}
						}
					}
				}
			}
		}()
	}
}

func (r *RegistryClient)GetConsulClient() *api.Client  {
	if r.Client != nil {
		return  r.Client
	}
	consulConfig :=JDSFGlobalConfig.Consul

	defaultConfig := api.DefaultConfig()

	consulPortStr := strconv.Itoa(int(consulConfig.Port))
	defaultConfig.Address = consulConfig.Address+":"+consulPortStr
	fmt.Println(defaultConfig.Address)
	defaultConfig.Scheme = consulConfig.Scheme
	client, err := api.NewClient(defaultConfig)
	if err != nil{
		fmt.Println(err)
		return nil
	}
	r.Client = client
	return client
}

func (r *RegistryClient)RegistryService()  {
	appConfig :=JDSFGlobalConfig.AppConfig
	ConsulDiscoverConfig := JDSFGlobalConfig.Consul.Discover
	portStr :=  strconv.Itoa(int(appConfig.ServerPort))
	client :=  r.Client
	if client == nil{
		client  = r.GetConsulClient()
	}
	agentService :=	new(api.AgentServiceRegistration)
	agentService.Port = int(appConfig.ServerPort)
	agentService.Address = appConfig.HostIp
	agentService.Kind = ""
	agentService.Name = appConfig.AppName
	agentService.ID = ConsulDiscoverConfig.ServiceInstanceId

	agentCheck :=new(api.AgentServiceCheck)
	agentCheck.Name =  appConfig.AppName
	agentCheck.CheckID = ConsulDiscoverConfig.ServiceInstanceId
	agentCheck.HTTP = "http://"+appConfig.HostIp+":"+portStr+ConsulDiscoverConfig.CheckUrl
	agentCheck.Method = "GET"
	agentCheck.Interval= "30s"
	agentService.Check = agentCheck
	regErr:=client.Agent().ServiceRegister(agentService)
	if regErr!=nil{
		fmt.Println(regErr)
	}
}



func (r *RegistryClient)ServiceRegistryCheck(serviceName string)  []*api.ServiceEntry{
	client := r.Client
	serviceEntry, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil
	}
	if len(serviceEntry) == 0 {
		fmt.Println("not found service name is ", serviceName)
		return nil
	}
	r.serviceCache.Put(serviceName,serviceEntry)
	return serviceEntry
}

func(r *RegistryClient)GetServiceInstance(serviceName string)[]*api.ServiceEntry {

	if serviceName == "" {
		return nil
	}
	val,ok := r.serviceCache.Load(serviceName)

    if !ok {
		return r.ServiceRegistryCheck(serviceName)
	}else {
		return val
	}

}

func (r *RegistryClient)LoadRegistryConfig(configType interface{}) interface{}  {
	client :=  r.Client
	if client == nil{
		client  = r.GetConsulClient()
	}
	configKey :=JDSFGlobalConfig.GetAppConsulConfigKey()
	keyPair,_,err := r.Client.KV().Get(configKey,nil)
	if err!=nil{
		log.Error().AnErr("get config form config server error",err)
	}
	yamlErr :=yaml.Unmarshal(keyPair.Value,configType)
	if yamlErr !=nil{
		log.Error().AnErr("format config from config center yaml to object err ",err)
		return nil
	}
	return configType
}

func (r *RegistryClient)ServiceRequestLoadBalance(rawURL string) string {

	reqURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println(err)
		return rawURL
	}
	serviceName := ""
	serviceNameAndPort := reqURL.Host

	serviceNameAndPortArray := strings.Split(serviceNameAndPort, ":")

	if len(serviceNameAndPortArray) > 0 {
		serviceName = serviceNameAndPortArray[0]
	}
	isMatch, err := regexp.MatchString("((?:(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)\\.){3}(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d))", serviceName)
	if isMatch {
		return rawURL
	}
	serviceEntry  := r.GetServiceInstance(serviceName);
	if serviceEntry == nil || len(serviceEntry) <= 0 {
		return rawURL
	}

	service := new(api.ServiceEntry)
	if len(serviceEntry) > 0 {
		rand.Seed(time.Now().UnixNano())
		serviceInstanceCount := len(serviceEntry)
		serviceIndex := rand.Intn(serviceInstanceCount)
		service = serviceEntry[serviceIndex]

	}
	if service.Service != nil {
		requestFinalHost := service.Service.Address + ":" + strconv.Itoa(service.Service.Port)
		reqURL.Host = requestFinalHost

		return reqURL.String()
	}
	return rawURL
}

