package logger_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/philips-labs/siderite/logger"
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
	control := make(chan string)
	marker := fmt.Sprintf("!^?%d?~!", uuid.New().ID())

	r, w, err := os.Pipe()
	if !assert.Nil(t, err) {
		return
	}
	err = logger.StartStorerWorker(r, ds, logging.Resource{
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
	}, control, marker)
	if !assert.Nil(t, err) {
		return
	}
	// Simulate task/func output
	quit := make(chan bool)
	go func() {
		for {
			_, _ = fmt.Fprintf(w, "%s\n", marker) // Immediate trigger
			select {
			case <-time.After(100 * time.Millisecond):
			case <-quit:
				return
			}
		}
	}()
	data := <-control
	quit <- true
	assert.Equal(t, marker, data)
}

func TestCustomLogEvent(t *testing.T) {
	testLog := `ERROR|CustomLogEvent|386a881c-de7a-4ba6-9acc-778e9897e997|a7629ac1d152517466d6ea17a499f3c4|4408ca86d9ce19d6|AuditClientResponseHandler|Error in persisting the audit message to audit repository`

	names := logger.CustomLogEventRegex.SubexpNames()
	md := map[string]string{}

	match := logger.CustomLogEventRegex.FindStringSubmatch(testLog)
	for i, n := range match {
		md[names[i]] = n
	}

	assert.Equal(t, "ERROR", md["severity"])
	assert.Equal(t, "4408ca86d9ce19d6", md["span_id"])
	assert.Equal(t, "a7629ac1d152517466d6ea17a499f3c4", md["trace_id"])
}
