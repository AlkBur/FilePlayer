package main

import (
	"os"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Database struct {
		File string `json:"file"`
		Path string `json:"mp3"`
	} `json:"database"`
	Host  string `json:"host"`
	Port  int `json:"port"`
	Debug bool `json:"debug"`
}

func LoadConfiguration(file string) *Config {
	var config = &Config{
		Host: "localhost",
		Port: 8081,
		Debug: false,
	}
	config.Database.File = "data.db"
	config.Database.Path = "mp3"

	configFile, err := os.Open(file)
	if err != nil {
		logDebug("Config: %s", err.Error())
		var b  []byte
		b, err = json.Marshal(config)
		if err == nil {
			err = ioutil.WriteFile(file, b, 0644)
			if err != nil {
				logDebug("Save config: %s", err.Error())
			}
		}
		return config
	}
	defer  configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(config)

	return config
}