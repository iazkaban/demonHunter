package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type HunterConfig struct {
	Server Server
	Login  Login
	System System
}
type Server struct {
	StartUrls []string
	UrlRules  []string
}

type System struct {
	GoroutineCounts int
	TimeSleep       int64
}

type Login struct {
	GetValues  map[string]string
	PostValues map[string]string
	LoginUrl   string
}

var Config HunterConfig

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
	fmt.Println(Config)
	if err != nil {
		return err
	}
	return nil
}
