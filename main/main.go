package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func init() {
	fmt.Println(colorRed + "Initializing songs" + colorReset)
	err := os.Chdir("songs")
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for _, file := range files {
		fileInfo, err := os.Stat(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		if !fileInfo.IsDir() {
			file, err := os.Open(file.Name())
			if err != nil {
				log.Fatal(err)
			}

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

func retrieveFile(name string) (song, error) {
	for _, song := range songs {

		// sets strings to lowercase and removes whitespaces so that the comparison is case-insensitive
		sTitle := strings.ToLower(strings.ReplaceAll(song.Title, " ", ""))
		name = strings.ToLower(strings.ReplaceAll(name, " ", ""))

		// check if the song exists in the database
		if strings.EqualFold(sTitle, name) {
			return song, nil
		}
	}

	return song{}, errors.New("song not found")
}

func getSongs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, songs)
}

func getSongByTitle(c *gin.Context) {
	title := c.Param("title")
	song, err := retrieveFile(title)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, song)
}

func readFileContents(song song) []byte {
	fname := "../songs/" + song.Title + "." + song.FileType
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	ctn, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return ctn
}

func playSong(c *gin.Context) {
	title := c.Param("title")
	song, err := retrieveFile(title)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "song not found"})
		return
	}

	connPool := audiopipeline.NewConnectionPool()

	go audiopipeline.Stream(connPool, readFileContents(song), float32(song.Length))

	w := c.Writer

	r := c.Request

	w.Header().Add("Content-Type", "audio/"+song.FileType)
	w.Header().Add("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)

	if !ok {
		log.Println("Could not create flusher")
	}

	connection := audiopipeline.MakeConnection()
	connPool.AddConnection(connection)
	log.Printf("%s has connected to the audio stream\n", r.Host)

	fileName := song.Title + "." + song.FileType
	fmt.Println("Playing: ", fileName)

	for {
		// Receive data from the buffer channel
		bufferChannel, buffer := audiopipeline.GetConnectionBuffers(connection)
		buf := <-bufferChannel
		if _, err := w.Write(buf); err != nil {
			connPool.DeleteConnection(connection)
			log.Printf("%s's connection to the audio stream has been closed\\n", r.Host)
			return
		}
		flusher.Flush()
		clear(buffer)
	}
}

func main() {
	router := gin.Default()
	router.GET("/songs", getSongs)
	router.GET("/songs/:title", getSongByTitle)
	router.GET("/play/:title", playSong)

	router.Run("localhost:8080")
}
