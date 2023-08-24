package logger

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/philips-labs/siderite/models"
	"github.com/philips-software/go-hsdp-api/logging"
)

var (
	CustomLogEventRegex = regexp.MustCompile(`^(?P<severity>[^\|\s]+)\s*\|\s*CustomLogEvent\s*\|\s*(?P<transaction_id>[^\|\s]*)\s*\|\s*(?P<trace_id>[^\|\s]*)\s*\|\s*(?P<span_id>[^\|\s]*)\s*\|\s*(?P<component_name>[^\|\s]*)\s*\|\s*(?P<logdata_message>.*)$`)
)

type Storer interface {
	StoreResources(messages []logging.Resource, count int) (*logging.StoreResponse, error)
}

func Setup(p models.Payload, taskID string) (chan string, string, func(), error) {
	var err error
	control := make(chan string)
	old := os.Stdout // keep backup of the real stdout
	marker := fmt.Sprintf("!^?%d?~!", uuid.New().ID())

	r, w, err := os.Pipe()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error setting up pipe: %v\n", err)
		return nil, marker, nil, err
	}
	os.Stdout = w

	var storer Storer

	// HSDP LogDrainer
	storer, err = NewLogDrainerStorer(p.Env)

	if err != nil {
		// HSDP Logging
		storer, err = NewHSDPStorer(p.Env)
	}
	if err != nil {
		os.Stdout = old // Reset
		return nil, marker, func() {}, err
	}
	err = StartStorerWorker(r, storer, logging.Resource{
		ResourceType:        "LogEvent",
		ApplicationInstance: taskID,
		EventID:             "1",
		ApplicationName:     "hsdp_function",
		ApplicationVersion:  "1.0.0",
		Component:           "siderite",
		Category:            "FunctionLog",
		Severity:            "info",
		OriginatingUser:     "siderite",
		ServerName:          "hsdp-function.siderite.ironworker",
		ServiceName:         taskID,
	}, control, marker)
	if err != nil {
		os.Stdout = old
		_, _ = fmt.Fprintf(os.Stderr, "[siderite] not logging to HSDP: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "[siderite] logging stdout to HSDP\n")
	}
	return control, marker, func() {
		os.Stdout = old
	}, nil
}

func StartStorerWorker(fd *os.File, client Storer, template logging.Resource, control chan string, marker string) error {
	fdReader := bufio.NewReader(fd)

	go func() {
		_, _ = fmt.Fprintf(os.Stderr, "[siderite] logging worker started\n")
		names := CustomLogEventRegex.SubexpNames()
		md := map[string]string{}
		for {
			text, err := fdReader.ReadString('\n')
			if err != nil {
				text = fmt.Sprintf("error reading: %v\n", err)
			}

			if strings.Contains(text, marker) {
				_, _ = fmt.Fprintf(os.Stderr, "[siderite] received control marker. exiting logger\n")
				control <- marker // Notify task/function runner
				return
			}

			// Prepare message
			msg := template
			msg.ID = uuid.New().String()
			msg.TransactionID = msg.ID
			msg.LogData.Message = base64.StdEncoding.EncodeToString([]byte(text))
			msg.LogTime = time.Now().Format("2006-01-02T15:04:05.000Z07:00")

			// Check for CustomLogEvent
			match := CustomLogEventRegex.FindStringSubmatch(text)
			for i, n := range match {
				md[names[i]] = n
			}
			if len(match) > 0 {
				msg.Severity = md["severity"]
				msg.TransactionID = md["transaction_id"]
				msg.TraceID = md["trace_id"]
			}

			// Check for passthrough events
			var pt logging.Resource
			if err := json.Unmarshal([]byte(text), &pt); err == nil && pt.ResourceType == "LogEvent" {
				msg = pt // Replace
			}

			if text != "" {
				resp, err := client.StoreResources([]logging.Resource{
					msg,
				}, 1)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "error storing: %v [%+v]\n", err, resp)
				}
			}
		}
	}()

	return nil
}
