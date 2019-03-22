package service

import (
	"encoding/json"
	"hash/crc32"
	"net/http"
)

type OperationResult struct {
	Status int `json:"status"`
	Message string `json:"message"`
	Result interface{} `json:"result"`
}

type GameInfoResult struct {
	GameId string `json:"gameId"`
	GameName string `json:"gameName"`
}
var(
	name  =  [5]string{"魔兽世界","守望先锋","英雄联盟","王者荣耀","炉石传说"}
)

func GetGameInfo(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()
	gameId := r.Form.Get("gameid")
	if gameId != "" {
		index := HashCode(gameId)%len(name)
		gameInfoResult := new (GameInfoResult)
		gameInfoResult.GameId = gameId
		gameInfoResult.GameName = name[index]
		operationResult := new (OperationResult)
		operationResult.Status = 200
		operationResult.Result = gameInfoResult
		data, _ := json.Marshal(operationResult)
		writeJsonResponse(w,200,data)
	}else{
		operationResult := new (OperationResult)
		operationResult.Status = 400
		operationResult.Message = "gameId can not be null"
		data, _ := json.Marshal(operationResult)
		writeJsonResponse(w,200,data)
	}


}

func HashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}