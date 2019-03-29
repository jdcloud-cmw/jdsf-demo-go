package service

import (
	"fmt"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/jdsfapi"
	"github.com/jdcloud-cmw/jdsf-demo-go/jdsf-demo-client/sling"
	"github.com/opentracing/opentracing-go"
	"io/ioutil"
	"net/http"
)

func RegistryCheck(w http.ResponseWriter, r *http.Request) {
	t1 := opentracing.GlobalTracer()
	fmt.Println(t1)
	paramMap := make(map[string]string)

	paramMap["gameid"] = "123123"

	t := opentracing.GlobalTracer()
	fmt.Println(t)
	req, err :=sling.New().Get("http://db-service:8090/db/gameinfo/getgameinfo?gameid=123123").EnableTrace(r.Context()).Request()

	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{Transport: &jdsfapi.Transport{}}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	writeJsonResponse(w, http.StatusOK, body)
}
