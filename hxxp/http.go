package hxxp

import (
	"encoding/json"
	"net/http"
)

const (
	RespCodeOK = 0
)

type M map[string]interface{}

type RespJSON struct {
	ErrCode    int         `json:"err_code"`
	ErrSubCode int         `json:"err_sub_code"`
	ErrMsg     string      `json:"err_msg"`
	Data       interface{} `json:"data,omitempty"`
}

func JSONOk(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private, max-age=0")
	err := json.NewEncoder(w).Encode(RespJSON{ErrCode: RespCodeOK, Data: data})
	return err
}
