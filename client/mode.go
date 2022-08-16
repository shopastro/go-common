package client

type (
	IResponse interface {
		SetBody(body string, code ...int) *Response
		GetHttpStatus() int
		GetBody() string
		GetStructByBody(data interface{}) error
	}
)
