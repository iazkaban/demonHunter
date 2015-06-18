package main

import (
	"fmt"
	"os"

	"github.com/iazkaban/demonHunter/config"
)

func main() {
	err := config.LoadConfigFile("./config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
