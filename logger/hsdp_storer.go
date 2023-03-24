package logger

import (
	"os"

	"github.com/philips-software/go-hsdp-api/iam"
	"github.com/philips-software/go-hsdp-api/logging"
)

const (
	LogIngestorKeyEnv               = "SIDERITE_LOGINGESTOR_KEY"
	LogIngestorSecretEnv            = "SIDERITE_LOGINGESTOR_SECRET"
	LogIngestorServiceIDEnv         = "SIDERITE_LOGINGESTOR_SERVICE_ID"
	LogIngestorServicePrivateKeyEnv = "SIDERITE_LOGINGESTOR_SERVICE_PRIVATE_KEY"
	LogIngestorRegionEnv            = "SIDERITE_LOGINGESTOR_REGION"
	LogIngestorEnvironmentEnv       = "SIDERITE_LOGINGESTOR_ENVIRONMENT"
	LogIngestorProductKeyEnv        = "SIDERITE_LOGINGESTOR_PRODUCT_KEY"
	LogIngestorURLEnv               = "SIDERITE_LOGINGESTOR_URL"
	LogIngestorDebug                = "SIDERITE_LOGINGESTOR_DEBUG"
)

func NewHSDPStorer(env map[string]string) (Storer, error) {
	var err error

	sharedKey := env[LogIngestorKeyEnv]
	sharedSecret := env[LogIngestorSecretEnv]
	serviceID := env[LogIngestorServiceIDEnv]
	servicePrivateKey := env[LogIngestorServicePrivateKeyEnv]
	region := env[LogIngestorRegionEnv]
	environment := env[LogIngestorEnvironmentEnv]

	cfg := &logging.Config{
		BaseURL:      env[LogIngestorURLEnv],
		ProductKey:   env[LogIngestorProductKeyEnv],
		SharedSecret: sharedSecret,
		SharedKey:    sharedKey,
		Region:       region,
		Environment:  environment,
	}
	if env[LogIngestorDebug] != "" {
		cfg.DebugLog = os.Stderr
	}
	client, err := logging.NewClient(nil, cfg)
	if err != nil {
		return nil, err
	}
	service := iam.Service{
		ServiceID:  serviceID,
		PrivateKey: servicePrivateKey,
	}
	if service.Valid() {
		err = client.ServiceLogin(service)
	}
	return client, err
}
