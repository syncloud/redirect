package main

import (
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("usage: ", os.Args[0], "/path.sock")
		return
	}

	config := utils.NewConfig()
	config.Load(os.Args[1])
	database := db.NewMySql()
	database.Connect(config.GetMySqlDB(), config.GetMySqlLogin(), config.GetMySqlPassword())
	api := rest.NewApi()
	api.Start(config.GetApiSocket())
}
