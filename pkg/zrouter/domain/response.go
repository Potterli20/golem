package domain

import (
	"encoding/json"
	"net/http"
	"sync"
)

const (
	ContentTypeHeader          = "Content-Type"
	ContentTypeApplicationJSON = "application/json; charset=utf-8"
	ContentTypePlainText       = "text/plain; charset=utf-8"
	ContentTypeJSON            = "json"
)

type ServiceResponse interface {
	Status() int
	Header() http.Header
	ResponseBytes() ([]byte, error)
	ResponseFormat() string
	Contents() any
}

type defaultServiceResponse struct {
	status        int
	header        http.Header
	response      any
	once          sync.Once
	responseBytes []byte
	marshalError  error
}

func (d *defaultServiceResponse) Status() int {
	return d.status
}

func (d *defaultServiceResponse) Header() http.Header {
	h := d.header
	if h == nil {
		h = http.Header{}
	}
	if h.Get(ContentTypeHeader) == "" {
		h.Set(ContentTypeHeader, ContentTypeApplicationJSON)
	}
	return h
}

func (d *defaultServiceResponse) ResponseFormat() string {
	return ContentTypeJSON
}

func (d *defaultServiceResponse) ResponseBytes() ([]byte, error) {
	d.once.Do(func() {
		if d.response != nil {
			d.responseBytes, d.marshalError = json.Marshal(d.response)
		} else {
			d.responseBytes = []byte{}
		}
	})
	return d.responseBytes, d.marshalError
}

func (d *defaultServiceResponse) Contents() any {
	return d.response
}

func NewServiceResponse(status int, response any) ServiceResponse {
	return &defaultServiceResponse{
		status:   status,
		response: response,
		header:   nil,
	}
}

func NewServiceResponseWithHeader(status int, response any, header http.Header) ServiceResponse {
	return &defaultServiceResponse{
		status:   status,
		response: response,
		header:   header,
	}
}

func NewErrorResponse(status int, errorCode, errMsg string) ServiceResponse {
	apiError := NewAPIErrorResponse(status, errorCode, errMsg)
	apiErrorBytes, err := json.Marshal(apiError)
	if err != nil {
		return NewServiceResponse(status, errMsg)
	}

	return &defaultServiceResponse{
		status:        status,
		response:      apiError,
		header:        nil,
		responseBytes: apiErrorBytes,
	}
}

func NewErrorNotFound(errMsg string) ServiceResponse {
	return NewErrorResponse(http.StatusNotFound, "ROUTE_NOT_FOUND", errMsg)
}
