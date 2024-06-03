package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tcolgate/mp3"
	"gorm.io/gorm"

	"infiniti.com/internal/audiopipeline"
	"infiniti.com/internal/database"
)

var db *gorm.DB

const PORT = "9000"

func checkErrorAndExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkErrorAndPass(err error) {
	if err != nil {
		log.Print(err)
	}
}

func checkErrorAndReturn(err error, c *gin.Context, message string) {
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": message})
		return
	}
}

func init() {
	err := godotenv.Load("/src/.env")
	checkErrorAndExit(err)

	dbHost := os.Getenv("DB_HOST")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	fmt.Println("Connecting to database...", dbHost, dbPass, dbName)

	wd, _ := os.Getwd()
	fmt.Println("Working directory:", wd)

	dir, _ := os.ReadDir(wd)
	fmt.Println("Files in working directory:", dir)

	db, err = database.Connect(dbHost, dbPass, dbName)
	checkErrorAndExit(err)
	database.Migrate(db)
	database.Seed(db, "../../resources/songs")
}

// source: https://stackoverflow.com/questions/60281655/how-to-find-the-length-of-mp3-file-in-golang
func calculatePlayLength(path string) float64 {
	t := 0.0

	r, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0

	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return 0.0
		}

		t = t + f.Duration().Seconds()
	}
	return t
}

func getSong(c *gin.Context) (database.Song, error) {
	param := c.Param("param")

	id, succes := strconv.Atoi(param)
	if succes == nil {
		song, err := database.GetSongById(db, id)
		if err == nil {
			return *song, nil
		}
		return database.Song{}, err
	} else {
		songs, err := database.GetSongByTitle(db, param)
		if err == nil && len(songs) > 0 {
			return songs[0], nil
		}
		err = fmt.Errorf("song not found")
		return database.Song{}, err
	}

}

func openFile(song database.Song) (*os.File, error) {
	fname := "../songs/" + song.Path
	file, err := os.Open(fname)
	checkErrorAndPass(err)

	return file, err
}

func readSongContents(file *os.File) ([]byte, error) {
	ctn, err := io.ReadAll(file)
	checkErrorAndPass(err)

	return ctn, err
}

func playSong(c *gin.Context) {
	song, err := getSong(c)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		return
	}

	connPool := audiopipeline.NewConnectionPool()

	file, err := openFile(song)
	checkErrorAndReturn(err, c, "Couldn't open file of song")

	bytes, err := readSongContents(file)
	checkErrorAndReturn(err, c, "Couldn't read song contents")

	// create a go routine to stream the song
	go audiopipeline.Stream(connPool, bytes, float32(calculatePlayLength("../songs/"+song.Path)))

	audiopipeline.PlayAudiofile(connPool, song.FileType, c)
}

func getSongs(c *gin.Context) {
	var songs = database.GetSongs(db)
	if songs != nil {
		c.IndentedJSON(http.StatusOK, songs)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Such empty"})
	}
}

func removeSong(c *gin.Context) {
	song, err := getSong(c)
	fmt.Println("song:", song, "err:", err)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
	} else {
		c.IndentedJSON(http.StatusOK, song)
		err = database.RemoveSong(db, song)
		if err != nil {
			log.Println(err)
		}
		database.RemoveSong(db, song)

	}
}

func uploadSong(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	checkErrorAndReturn(err, c, "file not found")

	err = c.SaveUploadedFile(fileHeader, "../songs/"+fileHeader.Filename)
	checkErrorAndReturn(err, c, "couldn't save file")

	file, err := os.Open(fileHeader.Filename)
	checkErrorAndReturn(err, c, "couldn't open file")

	database.ParseFileToSongDatatype(db, file)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", fileHeader.Filename))
}

func searchSong(c *gin.Context) {
	searchTerm := c.Param("param")
	songs, err := database.GetSongByTitle(db, searchTerm)
	checkErrorAndReturn(err, c, "song not found")

	c.IndentedJSON(http.StatusOK, songs)
}

func main() {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	router.GET("/songs", getSongs)
	router.GET("/search/:param", searchSong)

	// get song via title or id
	router.GET("/songs/:param", func(c *gin.Context) {
		song, err := getSong(c)

		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		} else {
			c.IndentedJSON(http.StatusOK, song)
		}
	})

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Infiniti! \n\nAvailable endpoints: \n\nGET /songs \nGET /songs/:param \nGET /search/:param \nGET /play/:param \nGET /remove/:param \nPOST /upload \n\nEnjoy!")
	})

	// play song via title or id
	router.GET("/play/:param", playSong)

	// remove sing by id or title
	router.GET("/remove/:param", removeSong)

	router.POST("/upload", uploadSong)
	router.Run("0.0.0.0:" + PORT)
}
