package logger

import (
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
	assert.Greater(d.t, 0, count)
	assert.Greater(d.t, 0, len(messages))
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

	err := toStorer(os.Stdout, ds, logging.Resource{
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

	assert.Nil(t, err)
}
