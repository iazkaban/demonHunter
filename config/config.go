package config

import (
	"encoding/json"
	"errors"
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
	UrlUnruly []string
	LimitHost []string
	//UrlRegExpRules []*regexp.Regexp
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
	if err != nil {
		return err
	}
	/*
		for i, v := range Config.Server.UrlRules {
			Config.Server.UrlRegExpRules[i] = regexp.MustCompile(v)
		}
	*/
	return nil
}
