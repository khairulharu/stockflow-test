package main

import (
	"fmt"
	"stockflow/internal/config"
)

func main() {
	viperConfig := config.NewViper()
	if viperConfig.GetString("app.name") != "stockflow" {
		panic("error loading config")
	}
	fmt.Println(viperConfig.GetString("app.name"))

	db := config.NewDatabase(viperConfig)
	if db != nil {
		fmt.Printf("connecting to database: %s \n", viperConfig.GetString("database.name"))
	}
}
