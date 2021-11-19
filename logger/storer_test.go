package logger

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/philips-software/go-hsdp-api/logging"
	"github.com/stretchr/testify/assert"
)

type dummyStorer struct {
	t *testing.T
}

func (d *dummyStorer) StoreResources(messages []logging.Resource, count int) (*logging.StoreResponse, error) {
	if !assert.Greater(d.t, 0, count) {
		return nil, fmt.Errorf("count is 0")
	}
	if !assert.Greater(d.t, 0, len(messages)) {
		return nil, fmt.Errorf("zero messages in slice")
	}
	return &logging.StoreResponse{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       nil,
		},
	}, nil
}

func TestToHSDP(t *testing.T) {
	ds := &dummyStorer{t}
	done := make(chan bool)

	err := startStorerWorker(os.Stdout, ds, logging.Resource{
		ResourceType:        "LogEvent",
		ApplicationInstance: "foo",
		EventID:             "1",
		ApplicationName:     "hsdp_function",
		ApplicationVersion:  "1.0.0",
		Component:           "siderite",
		Category:            "FunctionLog",
		Severity:            "info",
		OriginatingUser:     "siderite",
		ServerName:          "iron.io",
		ServiceName:         "foo",
	}, done)
	done <- true

	assert.Nil(t, err)
}
