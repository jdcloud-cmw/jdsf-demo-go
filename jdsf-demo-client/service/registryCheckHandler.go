package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/jdsfapi"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/sling"
	"github.com/opentracing/opentracing-go"
	"io/ioutil"
	"net/http"
	"strings"
)

func RegistryCheck(w http.ResponseWriter, r *http.Request) {
	t1 := opentracing.GlobalTracer()
	fmt.Println(t1)
 	r.ParseForm()
	gameid := r.Form.Get("gameid")
	if gameid == ""{
		gameid = "123123"
	}
	gameid = processRequestParam(gameid,r.Context());
	req, err :=sling.New().Get("http://db-service:8090/db/gameinfo/getgameinfo?gameid="+gameid).EnableTrace(r.Context()).Request()

	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{Transport: &jdsfapi.Transport{}}
	resp, err := client.Do(req)
	if err != nil{
		writeJsonResponse(w, http.StatusInternalServerError,   []byte("Internal Server Error"))
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	writeJsonResponse(w, http.StatusOK, body)
}

func processRequestParam(gameId string,reqContext context.Context) string  {

	gameId = gameId +strings.Replace(uuid.New().String(),"-","",-1)
	if jdsfapi.JDSFGlobalConfig.Tracing.Enable{
		var tr opentracing.Tracer
		var span opentracing.Span
		if reqContext != nil {
			span = opentracing.SpanFromContext(reqContext)
			tr = span.Tracer()
			currentSpan := tr.StartSpan("processRequestParam",opentracing.FollowsFrom(span.Context()))
			currentSpan.SetTag("test","testTag")
			currentSpan.Finish()
		}

	}

	return gameId
}
