package logger_test

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/go-syslog/v2/rfc5424"
	"github.com/philips-labs/siderite/logger"
	"github.com/philips-software/go-hsdp-api/logging"
	"github.com/stretchr/testify/assert"
)

var (
	muxLogger    *http.ServeMux
	loggerServer *httptest.Server
	parser       = rfc5424.NewParser()
)

func TestLogDrainer(t *testing.T) {
	teardown, err := setup(t)
	if !assert.Nil(t, err) {
		return
	}
	defer teardown()

	muxLogger.HandleFunc("/1234", endpointMocker(t, `{"message":"okay"}`, http.StatusOK))

	logDrainerURL := loggerServer.URL + "/1234"

	storer, err := logger.NewLogDrainerStorer(map[string]string{
		logger.LogDrainerEnv: logDrainerURL,
	})
	if !assert.Nil(t, err) {
		return
	}
	resp, err := storer.StoreResources([]logging.Resource{
		{
			ResourceType: "EventLog",
			LogTime:      time.Now().Format(time.RFC3339),
			LogData: logging.LogData{
				Message: base64.StdEncoding.EncodeToString([]byte("Hello world")),
			},
			ApplicationInstance: "app_instance",
			ApplicationName:     "hsdp_function",
			Severity:            "INFO",
			ServerName:          "test.terrakube.com",
		},
	}, 1)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, resp.Response) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func endpointMocker(t *testing.T, responseBody string, statusCode ...int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := ioutil.ReadAll(r.Body)

		_, err := parser.Parse(body)
		if !assert.Nil(t, err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(statusCode) > 0 {
			w.WriteHeader(statusCode[0])
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write([]byte(responseBody))
	}
}

func setup(t *testing.T) (func(), error) {
	var err error

	muxLogger = http.NewServeMux()
	loggerServer = httptest.NewServer(muxLogger)

	if err != nil {
		return func() {
			loggerServer.Close()
		}, err
	}

	return func() {
		loggerServer.Close()
	}, nil
}
