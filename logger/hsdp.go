package logger

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/philips-software/go-hsdp-api/logging"
)

func ToHSDP(fd *os.File, template logging.Resource, done chan bool) error {
	client, err := logging.NewClient(http.DefaultClient, &logging.Config{
		SharedKey:    os.Getenv("SIDERITE_LOGGING_SHARED_KEY"),
		SharedSecret: os.Getenv("SIDERITE_LOGGING_SHARED_SECRET"),
		BaseURL:      os.Getenv("SIDERITE_LOGGING_BASE_URL"),
		ProductKey:   os.Getenv("SIDERITE_LOGGING_PRODUCT_KEY"),
	})
	if err != nil {
		return err
	}
	fdReader := bufio.NewReader(fd)
	go func() {
		for {
			// Next line
			text, _ := fdReader.ReadString('\n')
			//fmt.Fprintf(os.Stderr, "Logging: %s\n", text)
			// Prepare message
			template.ID = uuid.New().String()
			template.TransactionID = template.ID
			template.LogData.Message = text
			template.LogTime = time.Now().Format("2006-01-02T15:04:05.000Z07:00")

			resp, err := client.StoreResources([]logging.Resource{
				template,
			}, 1)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error storing: %v [%v]\n", err, resp)
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
