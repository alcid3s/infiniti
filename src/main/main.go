package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"infiniti/audiopipeline"
	"infiniti/database"

	"github.com/gin-gonic/gin"
	"github.com/tcolgate/mp3"
	"gorm.io/gorm"
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
	file, err := os.Open("credentials.txt")
	checkErrorAndExit(err)

	contents, err := io.ReadAll(file)
	checkErrorAndExit(err)

	credentials := strings.Split(string(contents), ":")

	db, err = database.Connect(credentials[0], credentials[1], credentials[2])
	checkErrorAndExit(err)
	database.Migrate(db)
	database.Seed(db, "../songs")
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
	checkErrorAndReturn(err, c, "song not found")

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
	checkErrorAndExit(err)
	database.RemoveSong(db, song)
}

func uploadSong(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	checkErrorAndExit(err)

	err = c.SaveUploadedFile(fileHeader, "../songs/"+fileHeader.Filename)
	checkErrorAndExit(err)

	file, err := os.Open(fileHeader.Filename)
	checkErrorAndExit(err)

	database.ParseFileToSongDatatype(db, file)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", fileHeader.Filename))
}

func main() {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	router.GET("/songs", getSongs)

	// get song via title or id
	router.GET("/songs/:param", func(c *gin.Context) {
		song, err := getSong(c)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		} else {
			c.IndentedJSON(http.StatusOK, song)
		}
	})

	// play song via title or id
	router.GET("/play/:param", playSong)

	// remove sing by id or title
	router.GET("/remove/:param", removeSong)

	router.POST("/upload", uploadSong)
	router.Run("0.0.0.0:" + PORT)
}
