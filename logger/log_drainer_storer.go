package logger

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

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
		decoded, err := base64.StdEncoding.DecodeString(msg.LogData.Message)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		syslogMessage := fmt.Sprintf("<14>1 %s %s %s %s - - %s", msg.LogTime, msg.ServerName, msg.ApplicationName, msg.ApplicationInstance, string(decoded))
		resp, err = l.Client.Post(l.logDrainerURL, "text/syslog", bytes.NewBufferString(syslogMessage))
		if err != nil {
			errs = append(errs, err)
		}
		if resp == nil || resp.StatusCode != http.StatusOK {
			_, _ = fmt.Fprintf(os.Stderr, "failed to send log: %v %v", resp, err)
		}
	}
	logResponse.Response = &http.Response{StatusCode: http.StatusOK}
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
