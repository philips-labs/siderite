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
	data, err_home := ioutil.ReadFile(configFile) //home
	local_data, err_local := ioutil.ReadFile("iron.json") //iron.json in current directory

	//If neither home or local json is found
	if err_local != nil && err_home != nil{
		return nil, err_home
	}

	var config Config
	if (err_home == nil){
		err_home = json.Unmarshal(data, &config)

		if(err_home != nil){
			return nil, err_home
		}
	}	
	if (err_local == nil){
		err_local = json.Unmarshal(local_data, &config)

		if(err_local != nil){
			return nil, err_local
		}
	}

	return &config, nil
}
