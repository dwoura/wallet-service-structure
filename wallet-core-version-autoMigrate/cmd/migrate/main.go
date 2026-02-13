package main

import (
	"flag"
	"fmt"
	"log"

	"wallet-core/pkg/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var command string
	flag.StringVar(&command, "cmd", "up", "Command to run: up, down")
	flag.Parse()

	// 加载配置
	config.Init()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Host,
		config.Global.DB.Port,
		config.Global.DB.Name,
	)

	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	if command == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up done")
	} else if command == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down done")
	} else {
		log.Fatalf("Unknown command: %s", command)
	}
}
