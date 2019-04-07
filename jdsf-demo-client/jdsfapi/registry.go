package jdsfapi

import (
	"fmt"
	"github.com/apex/log"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/jdsfapi/util"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)
const	consulHostEnv = "CONSUL_HOST"
const 	consulPortEnv  ="CONSUL_PORT"


type RegistryClient struct {
	Address string
	Port int
	Scheme string
	Client  *api.Client
	ServiceCache map[string][]*api.ServiceEntry
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
	JDSFRegistryClient = registryClient
	return registryClient
}

func (r *RegistryClient)serviceInfoTTL()  {

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



func (r *RegistryClient)ServiceRegistryCheck(serviceName string)  {

}

func (r *RegistryClient)ServiceRequestLoadBalance(rawURL string) string {
	if r.ServiceCache == nil{
		r.ServiceCache = make(map[string][]*api.ServiceEntry)
	}
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
		if r.ServiceCache[serviceName]!=nil  && len(r.ServiceCache[serviceName])>0{
			var index = 0
			indexUUID,uuidErr :=	uuid.NewUUID()
			if uuidErr!= nil{
				log.Errorf("get uuid index error",uuidErr)
			}
			hashCodeInt := util.HashCode(indexUUID.String())
			index = hashCodeInt % len(r.ServiceCache[serviceName])
			service := r.ServiceCache[serviceName][index]
			requestFinalHost := service.Service.Address + ":" + strconv.Itoa(service.Service.Port)
			reqURL.Host = requestFinalHost
			return reqURL.String()
		}
	}


	isMatch, err := regexp.MatchString("((?:(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)\\.){3}(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d))", serviceName)
	if isMatch {
		return rawURL
	}
	client := r.Client
	serviceEntry, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return rawURL
	}
	if len(serviceEntry) == 0 {
		fmt.Println("not found service name is ", serviceName)
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

