package logger

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/philips-software/go-hsdp-api/logging"
)

const (
	LogDrainerEnv = "SIDERITE_LOGDRAINER_URL"
)

type logDrainerStorer struct {
	logDrainerURL string
	*http.Client
}

func (l *logDrainerStorer) StoreResources(messages []logging.Resource, count int) (*logging.StoreResponse, error) {
	var errs []error
	var resp *http.Response
	logResponse := &logging.StoreResponse{}

	for i := 0; i < count; i++ {
		var err error
		msg := messages[i]
		structuredData := fmt.Sprintf("[spanId=\"%s\" traceId=\"%s\"]", msg.SpanID, msg.TraceID)
		syslogMessage := fmt.Sprintf("<14>1 %s %s %s %s %s %s", msg.LogTime, msg.ServerName, msg.ApplicationName, msg.ApplicationInstance, structuredData, msg.LogData.Message)
		resp, err = l.Client.Post(l.logDrainerURL, "text/syslog", bytes.NewBufferString(syslogMessage))
		if err != nil {
			errs = append(errs, err)
		}
		if resp == nil || resp.StatusCode != http.StatusOK {
			logResponse.Failed[i] = messages[i]
		}
	}
	logResponse.Response = resp
	return logResponse, nil
}

func NewLogDrainerStorer(env map[string]string) (Storer, error) {
	logDrainerURL := env[LogDrainerEnv]
	if logDrainerURL == "" {
		return nil, fmt.Errorf("missing '%s' needed by LogDrainerStorer", LogDrainerEnv)
	}
	storer := &logDrainerStorer{
		Client:        &http.Client{},
		logDrainerURL: logDrainerURL,
	}

	return storer, nil
}
