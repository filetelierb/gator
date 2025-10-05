package config

import (
	"encoding/json"
	"os"
	"fmt"
)

const configFileName = ".gatorconfig.json"


type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}
func getConfigFilePath() (string,error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + "/" + configFileName, nil

}


func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	file, err := os.Open(filePath)
	if err != nil{
		return Config{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var newConfig Config
	if err = decoder.Decode(&newConfig); err != nil{
		return Config{}, err
	}
	return newConfig, nil

}

func (config Config) SetUser(username string) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, os.ModePerm)
	if err != nil{
		return err
	}
	defer file.Close()
	newContent := []byte(fmt.Sprintf("{\"db_url\":\"%s\",\"current_user_name\":\"%s\"}",config.DbUrl,username))
	if _,err = file.Write(newContent); err != nil {
		return err
	}

	return nil


}