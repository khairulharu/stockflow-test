package main

import (
	"fmt"
	"stockflow/internal/config"
)

func main() {
	config := config.NewViper()
	if config.GetString("app.name") != "stockflow" {
		panic("error loading config")
	}
	fmt.Println(config.GetString("app.name"))
}
