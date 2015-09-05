package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/iazkaban/demonHunter/config"
	"github.com/iazkaban/demonHunter/contentanalyzer"
	"github.com/iazkaban/demonHunter/login"
)

var wg sync.WaitGroup

func main() {
	err := config.LoadConfigFile("./config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = login.Login()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := 0; i < len(config.Config.Server.StartUrls); i++ {
		contentanalyzer.SetUrl(config.Config.Server.StartUrls[i])
	}
	wg.Add(1)
	contentanalyzer.Run()
	wg.Wait()
}
