package jdsfapi

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func AppTraceGlobalConfig()  {
	appConfig := JDSFGlobalConfig.AppConfig
	traceConfig := JDSFGlobalConfig.Tracing
	fmt.Println(appConfig)
	trace ,_ := newTrace(appConfig.AppName,traceConfig)

	opentracing.SetGlobalTracer(trace)
	//defer  traceCloser.Close()
}

// New returns a new tracer
func newTrace(serviceName string,traceConfig *TraceConfig) (opentracing.Tracer, io.Closer) {
	var cfg config.Configuration
	var reporter  config.ReporterConfig
	if traceConfig!= nil && traceConfig.TraceUdpAddress!=""{
		agentHostPort := traceConfig.TraceUdpAddress+":"+strconv.Itoa(traceConfig.TraceUdpPort)
		fmt.Println(agentHostPort)
		reporter  = config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "10.12.140.173:5775", // localhost:5775
		}
	}else if traceConfig != nil && traceConfig.TraceHttpAddress !="" {
		httpHostPort := "http://" + traceConfig.TraceHttpAddress+":"+strconv.Itoa(traceConfig.TraceHttpPort)+"/api/traces"
		reporter = config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			CollectorEndpoint: httpHostPort,
		}
	}
	fmt.Println(reporter)
	simpleType := ""
	var param float64 = 1
	switch strings.ToLower(traceConfig.SimpleType) {
	case "const":
		simpleType="const"
		break
	case "probabilistic":
		simpleType="probabilistic"
		param = traceConfig.SimpleRate
		break
	case "rateLimiting":
		simpleType="rateLimiting"
		param = traceConfig.SimpleRate
		break
	case "remote":
		simpleType="remote"
		param = traceConfig.SimpleRate
		break
	default:
		simpleType="probabilistic"
		param = 0.1
	}
	cfg = config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  simpleType,
			Param: param,
		},
		Reporter: &reporter,
	}
	tracer, closer, err := cfg.New(
		serviceName,
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		log.Fatal(err)
	}

	return tracer, closer
}
