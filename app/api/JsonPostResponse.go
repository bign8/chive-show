package api

import (
  "encoding/json"
  "fmt"
  "net/http"
)

var successCodes map[int]bool = map[int]bool{
  200: true, // 200 OK
}

type JsonResponse struct {
  Status string      `json:"status"`
  Code   int         `json:"code"`
  Data   interface{} `json:"data"`
  Msg    string      `json:"msg"`
}

func NewJsonResponse(code int, message string, data interface{}) JsonResponse {
  status := "error"
  if successCodes[code] { status = "success" }
  return JsonResponse{
    Code:   code,
    Data:   data,
    Msg:    message,
    Status: status,
  }
}

func (res *JsonResponse) write(w http.ResponseWriter) error {
  str_items, err := json.MarshalIndent(&res, "", "  ")
  var out string
  if err != nil {
    out = "{\"status\":\"error\",\"code\":500,\"data\":null,\"msg\":\"Error marshaling data\"}"
  } else {
    out = string(str_items)
  }
  fmt.Fprint(w, out)
  return err
}
