package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	RedisHost string
	MysqlHost string
	IconUrl   string
}

func LoadConfig(filename string) (Configuration, error) {
	var config = Configuration{}
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
