package config

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"

	"github.com/lib/pq"
)

func NewDatabase(viper *viper.Viper) *sql.DB {
	cfg := pq.Config{
		Host:           viper.GetString("database.host"),
		Port:           uint16(viper.GetInt("database.port")),
		User:           viper.GetString("database.username"),
		Password:       viper.GetString("database.password"),
		Database:       viper.GetString("database.name"),
		SSLMode:        pq.SSLModeDisable,
		ConnectTimeout: 5 * time.Second,
	}

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		log.Fatalf("database: error create config, %v", err)
	}

	db := sql.OpenDB(c)

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		log.Fatalf("database: error to connect database, %v", err)
	}

	return db
}
