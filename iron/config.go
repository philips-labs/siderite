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

// Load loads the Iron config as found in ~/.iron.json
func Load(configFiles ...string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(home, ".iron.json")
	if len(configFiles) > 0 {
		configFile = configFiles[0]
	}
	data, errHome := ioutil.ReadFile(configFile)        //home
	localData, errLocal := ioutil.ReadFile("iron.json") //iron.json in current directory

	//If neither home or local json is found
	if errLocal != nil && errHome != nil {
		return nil, errHome
	}

	var config Config
	if errHome == nil {
		errHome = json.Unmarshal(data, &config)

		if errHome != nil {
			return nil, errHome
		}
	}
	if errLocal == nil {
		errLocal = json.Unmarshal(localData, &config)

		if errLocal != nil {
			return nil, errLocal
		}
	}

	return &config, nil
}
