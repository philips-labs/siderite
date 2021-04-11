package siderite

// Payload describes the siderite payload format
type Payload struct {
	Version  string            `json:"version"`
	Env      map[string]string `json:"env,omitempty"`
	Cmd      []string          `json:"cmd,omitempty"`
	Type     string            `json:"type"`
	Token    string            `json:"token,omitempty"`
	Auth     string            `json:"auth,omitempty"`
	Upstream string            `json:"upstream,omitempty"`
	Mode     string            `json:"mode,omitempty"`
}

// CronPayload describes cron payload stored on in Iron schedule payload field
type CronPayload struct {
	Schedule         string `json:"schedule"`
	EncryptedPayload string `json:"encrypted_payload"`
	Timeout          int    `json:"timeout,omitempty"`
}
