package client

import (
	"encoding/json"
	"golang.org/x/net/context"
	"net/http"
)

type (
	Response struct {
		Code       int    `json:"code"`
		body       string `json:"-"`
		httpStatus int    `json:"-"`
	}

	IResp struct {
	}
)

func NewResponse(ctx context.Context) *Response {
	return &Response{}
}

func (svc *Response) SetBody(body string, code ...int) *Response {
	if body == "" {
		return svc
	}

	status := http.StatusBadGateway
	if len(code) == 1 {
		status = code[0]
	}

	json.Unmarshal([]byte(body), &svc)

	svc.body = body
	svc.httpStatus = status

	return svc
}

func (svc *Response) GetHttpStatus() int {
	return svc.httpStatus
}

func (svc *Response) GetBody() string {
	return svc.body
}

func (svc *Response) GetStructByBody(data interface{}) error {
	err := json.Unmarshal([]byte(svc.body), &data)
	if err != nil {
		return err
	}

	return nil
}
