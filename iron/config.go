package iron

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config holds configuration of a HSDP Iron service
type Config struct {
	ClusterInfo []struct {
		ClusterID   string `json:"cluster_id"`
		ClusterName string `json:"cluster_name"`
		Pubkey      string `json:"pubkey"`
		UserID      string `json:"user_id"`
	} `json:"cluster_info"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Project   string `json:"project"`
	ProjectID string `json:"project_id"`
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
}

// LoadConfig loads the Iron config as found in ~/.iron.json
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(home, ".iron.json")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
