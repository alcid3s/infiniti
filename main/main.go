package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"CNDAPI/audiopipeline"

	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"
	"github.com/tcolgate/mp3"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

type song struct {
	Id       int     `json:"id"`
	Title    string  `json:"title"`
	FileType string  `json:"fileType"`
	Artist   string  `json:"artist"`
	Length   float64 `json:"length"`
}

var songs []song

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	fmt.Println(colorRed + "Initializing songs" + colorReset)
	err := os.Chdir("songs")
	checkError(err)

	files, err := os.ReadDir(".")
	checkError(err)

	i := 0
	for _, file := range files {
		fileInfo, err := os.Stat(file.Name())
		checkError(err)

		if !fileInfo.IsDir() {
			file, err := os.Open(file.Name())
			checkError(err)

			m, err := tag.ReadFrom(file)
			if err == nil {
				songs = append(songs, song{
					Id:       i,
					Title:    strings.Split(file.Name(), ".")[0],
					FileType: strings.Split(file.Name(), ".")[1],
					Artist:   m.Artist(),
					Length:   calculateTrackLength(file),
				})
			}
		}
		i++
	}
}

// source: https://stackoverflow.com/questions/60281655/how-to-find-the-length-of-mp3-file-in-golang
func calculateTrackLength(file *os.File) float64 {
	d := mp3.NewDecoder(file)
	var f mp3.Frame
	skipped := 0
	t := 0.0
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

	if t == 0.0 {
		log.Print(colorRed + "Couldn't calculate track length" + colorReset)
	}
	return t
}

func retrieveFile(name string, id int) (song, error) {
	toLower := func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", ""))
	}

	for _, song := range songs {
		if name != "" && toLower(song.Title) == toLower(name) {
			return song, nil
		}
		if id != -1 && song.Id == id {
			return song, nil
		}
	}

	return song{}, errors.New("song not found")
}

func getSongs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, songs)
}

func getSong(c *gin.Context) song {
	param := c.Param("param")

	var song song
	if id, err := strconv.Atoi(param); err == nil {
		song, err = retrieveFile("", id)
		checkError(err)
	} else {
		song, err = retrieveFile(param, -1)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
			checkError(err)
		}
	}

	return song
}

func readSongContents(song song) []byte {
	fname := "../songs/" + song.Title + "." + song.FileType
	file, err := os.Open(fname)
	checkError(err)

	ctn, err := io.ReadAll(file)
	checkError(err)

	return ctn
}

func playSong(c *gin.Context) {
	song := getSong(c)

	connPool := audiopipeline.NewConnectionPool()

	go audiopipeline.Stream(connPool, readSongContents(song), float32(song.Length))

	audiopipeline.PlayAudiofile(connPool, song.FileType, c)
}

func main() {
	router := gin.Default()
	router.GET("/songs", getSongs)

	// get song via title or id
	router.GET("/songs/:param", func(c *gin.Context) {
		song := getSong(c)
		c.IndentedJSON(http.StatusOK, song)
	})

	// play song via title or id
	router.GET("/play/:param", playSong)

	router.Run("localhost:8080")
}
