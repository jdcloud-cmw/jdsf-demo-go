package service

import "net/http"

type Route struct {
	Name 	string // 方法名称
	Method 	string // 路径
	Pattern string  // 路由适配
	HandlerFunc http.HandlerFunc // 调用方法
}

type Routies = []Route


var routes = Routies{
	Route{
		"HealthCheck",
		"GET",
		"/api/health/check",
		HealthCheck,
	},
	Route{
		"RegistryCheck",
		"GET",
		"/api/registry",
		RegistryCheck,
	},
}
