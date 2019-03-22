# JDSF Golang Demo

## 项目说明

* 本项目为JDSF Golang demo 项目，主要介绍如何使用京东JDSF框架如何构建一个简单的分布式服务。

## 环境准备

* 安装GoLang SDK,开发使用的是 `go1.10.4` 配置好`GOPATH` `GOROOT` 等相应的环境变量。

* 从此github 仓库 clone 代码 放到 `$GOPATH/github.com/jdcloud-cmw/` 下

* 安装自己喜欢的IDE 或者 vim 进行代码编辑

## 项目结构

|- jdsf-demo-client  
|- jdsf-demo-server  
|- README.md  
其中jdsf-demo-client为服务的消费者  
jdsf-demo-server 为服无的生产者  
README.md 为此文件
在 项目中 jdsfapi 为进行负载调用以及调用链注册的相关代码，[sling](https://github.com/dghubble/sling) 为 github上的开源项目，本项目在 sling 的基础上进行了扩展，所以直接引用了代码。  
项目中的 conf 文件夹存放了项目的配置  
service 文件夹为项目的启动 和 相关服务的逻辑

## 项目依赖类库说明

项目需要使用 go get 引用响应的依赖库，具体的如下

```shell
go get github.com/opentracing/opentracing-go   # opentracing 调用库
go get github.com/gorilla/mux  #http mux 实现  
go get github.com/google/uuid   #go lang uuid 实现库
go get gopkg.in/yaml.v2   #go yaml 实现库
go get github.com/uber/jaeger-client-go   # jaeger client 实现库
go get github.com/hashicorp/consul/api   # consul api 库
```

## 配置及使用说明

* 配置文件在项目的 conf 目录下的 appConfig.yaml 文件，具体的说明如下：  

```yaml
consul:
  scheme: http  # consul 使用协议
  address: 10.12.209.43 # consul 地址
  port: 8500 #consul 使用端口号
  discover:
    enable: true # 是否使用服务注册发现
    instanceId: go-consul-demo-1 # 服务注册的 instance id
    checkUrl: /api/health/check # 服务的健康检查地址
trace:
  enable: true  # 是否启用调用链
  simpleType: const # 调用链的采集模式
  simpleRate: 1 # 调用链的采集率
  traceHttpAddress: 10.12.140.173 # 调用链地址
  traceHttpPort: 14268 #调用链端口号
app:
  appName: go-consul-demo # 应用名称
  hostIp: 10.12.140.173 # 应用 hostIp
  serverPort: 19200 # 应用的端口号
```

## 代码运行及调试

* 在运行前使用 执行上面 的依赖包引用

* 在文件夹下执行   `go run main.go` 或者 执行 `go build` 后执行生成的可执行文件