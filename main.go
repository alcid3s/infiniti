package main

import (
	"fmt"
	"log"

	"infiniti.com/config"
	"infiniti.com/routes"

	song_controller "infiniti.com/controller"
	_ "infiniti.com/docs"
	"infiniti.com/internal/database"
)

func init() {
	dbHost := config.DB_HOST
	dbPass := config.DB_PASS
	dbName := config.DB_NAME

	fmt.Println("Connecting to database...", dbHost, dbPass, dbName)

	db, err := database.Connect(dbHost, dbPass, dbName)
	if err != nil {
		log.Fatal("Error connecting to database, err: ", err)
	}

	database.Migrate(db)
	database.Seed(db, "./resources/songs")
	song_controller.Init(db)
}

// @title Infiniti API
// @version 1.0
// @description This is a simple API for a music streaming service.
// @host 127.0.0.1:9000
func main() {
	router := routes.SetupRouter()

	router.Run("0.0.0.0:" + config.PORT)
}
