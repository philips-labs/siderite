package logger

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/philips-labs/siderite/models"
	"github.com/philips-software/go-hsdp-api/logging"
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
	err = startStorerWorker(r, storer, logging.Resource{
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

func startStorerWorker(fd *os.File, client Storer, template logging.Resource, control chan string, marker string) error {
	fdReader := bufio.NewReader(fd)

	go func() {
		_, _ = fmt.Fprintf(os.Stderr, "[siderite] logging worker started\n")
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
			template.ID = uuid.New().String()
			template.TransactionID = template.ID
			template.LogData.Message = base64.StdEncoding.EncodeToString([]byte(text))
			template.LogTime = time.Now().Format("2006-01-02T15:04:05.000Z07:00")

			if text != "" {
				resp, err := client.StoreResources([]logging.Resource{
					template,
				}, 1)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "error storing: %v [%v]\n", err, resp)
				}
			}
		}
	}()

	return nil
}
