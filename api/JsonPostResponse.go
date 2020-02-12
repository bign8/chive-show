package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var successCodes = map[int]bool{
	200: true, // 200 OK
}

// JSONResponse a wrapper for default responses
type JSONResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
}

// NewJSONResponse generate a JSONResponse
func NewJSONResponse(code int, message string, data interface{}) JSONResponse {
	status := "error"
	if successCodes[code] {
		status = "success"
	}
	return JSONResponse{
		Code:   code,
		Data:   data,
		Msg:    message,
		Status: status,
	}
}

func (res *JSONResponse) write(w http.ResponseWriter) error {
	items, err := json.MarshalIndent(&res, "", "  ")
	var out string
	if err != nil {
		out = "{\"status\":\"error\",\"code\":500,\"data\":null,\"msg\":\"Error marshaling data\"}"
	} else {
		out = string(items)
	}
	fmt.Fprint(w, out)
	return err
}
