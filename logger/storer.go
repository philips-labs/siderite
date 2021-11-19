package logger

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/philips-labs/siderite/models"
	"github.com/philips-software/go-hsdp-api/logging"
)

type Storer interface {
	StoreResources(messages []logging.Resource, count int) (*logging.StoreResponse, error)
}

func Setup(p models.Payload, taskID string) (chan bool, func(), error) {
	var err error
	done := make(chan bool)
	old := os.Stdout // keep backup of the real stdout
	r, w, err := os.Pipe()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error setting up pipe: %v\n", err)
		return done, nil, err
	}
	os.Stdout = w

	var storer Storer

	// HSDP LogDrainer
	storer, err = NewHLogDrainerStorer(p.Env)

	if err != nil {
		// HSDP Logging
		storer, err = NewHSDPStorer(p.Env)
	}
	if err != nil {
		return done, func() {}, err
	}
	err = toStorer(r, storer, logging.Resource{
		ResourceType:        "LogEvent",
		ApplicationInstance: taskID,
		EventID:             "1",
		ApplicationName:     "hsdp_function",
		ApplicationVersion:  "1.0.0",
		Component:           "siderite",
		Category:            "FunctionLog",
		Severity:            "info",
		OriginatingUser:     "siderite",
		ServerName:          "iron.io",
		ServiceName:         taskID,
	}, done)
	if err != nil {
		os.Stdout = old
		_, _ = fmt.Fprintf(os.Stdout, "[siderite] not logging to HSDP: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "[siderite] logging stdout to HSDP\n")
	}
	return done, func() {
		os.Stdout = old
		fmt.Printf("flushing logs\n")
		time.Sleep(3 * time.Second)
	}, nil
}

func toStorer(fd *os.File, client Storer, template logging.Resource, done chan bool) error {
	fdReader := bufio.NewReader(fd)
	go func() {
		for {
			// Next line
			text, _ := fdReader.ReadString('\n')
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
			// Check if we should exit
			select {
			case <-done:
				_, _ = fmt.Fprintf(os.Stderr, "exiting logger\n")
				return
			default:
			}
		}
	}()

	return nil
}
