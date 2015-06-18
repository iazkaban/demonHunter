package config

import (
	"encoding/json"
	"errors"
	"os"
)

type HunterConfig struct {
	Server Server
	Login  Login
}
type Server struct {
	Url    []string
	RegExp []string
}

type Login struct {
	Values LoginValues
	Url    string
}

type LoginValues struct {
	Url  string
	Get  map[string]string
	Post map[string]string
}

var Config HunterConfig

func init() {
}

func LoadConfigFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	content := make([]byte, 2048)
	len, err := file.Read(content)
	if err != nil {
		return err
	}
	if len == 0 {
		return errors.New("config file is empty")
	}
	err = json.Unmarshal(content[:len], &Config)
	if err != nil {
		return err
	}
	return nil
}
