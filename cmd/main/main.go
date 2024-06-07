package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"infiniti.com/internal/database"
	song_handler "infiniti.com/pkg/handles"
	"infiniti.com/pkg/routes"
)

const PORT = "9000"

func init() {
	err := godotenv.Load("/src/.env")
	if err != nil {
		log.Fatal("Error loading .env file, err: ", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	fmt.Println("Connecting to database...", dbHost, dbPass, dbName)

	db, err := database.Connect(dbHost, dbPass, dbName)
	if err != nil {
		log.Fatal("Error connecting to database, err: ", err)
	}

	database.Migrate(db)
	database.Seed(db, "../../resources/songs")
	song_handler.Init(db)
}

func main() {
	router := routes.SetupRouter()
	router.Run("0.0.0.0:" + PORT)
}
