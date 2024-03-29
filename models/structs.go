package models

// Payload describes the siderite payload format
type Payload struct {
	Version  string            `json:"version"`
	Name     string            `json:"name,omitempty"`
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
	Name             string `json:"name,omitempty"`
	EncryptedPayload string `json:"encrypted_payload"`
	Type             string `json:"type,omitempty"`
	Timeout          int    `json:"timeout,omitempty"`
}
