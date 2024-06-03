/*
---------------------------------------------------------------------------------------------------------------------------
	The sourcecode for the audiopipeline has been greatly inspired by: https://github.com/Icelain/radio/blob/main/main.go
	Many thanks to the author of the original code!
---------------------------------------------------------------------------------------------------------------------------
*/

package audiopipeline

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	BUFFERSIZE = 16384

	// hardcoded delay for if track length hasn't been provided correctly
	DELAY         = 150
	PRODUCTFACTOR = 8
)

type Connection struct {
	bufferChannel chan []byte
	buffer        []byte
}

type ConnectionPool struct {
	ConnectionMap map[*Connection]struct{}
	mu            sync.Mutex
}

func (cp *ConnectionPool) AddConnection(connection *Connection) {
	defer cp.mu.Unlock()
	cp.mu.Lock()
	cp.ConnectionMap[connection] = struct{}{}
}

func (cp *ConnectionPool) DeleteConnection(connection *Connection) {
	defer cp.mu.Unlock()
	cp.mu.Lock()
	delete(cp.ConnectionMap, connection)
}

func (cp *ConnectionPool) Broadcast(buffer []byte) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	for connection := range cp.ConnectionMap {
		copy(connection.buffer, buffer)
		select {
		case connection.bufferChannel <- connection.buffer:
		default:
		}
	}
}

func NewConnectionPool() *ConnectionPool {
	connectionMap := make(map[*Connection]struct{})
	return &ConnectionPool{ConnectionMap: connectionMap}
}

func Stream(connectionPool *ConnectionPool, content []byte, track_length float32) {
	buffer := make([]byte, BUFFERSIZE)

	for {
		clear(buffer)
		tempfile := bytes.NewReader(content)

		ticker := time.NewTicker(time.Millisecond * DELAY)

		if track_length != 0.0 {
			timer := time.Duration(float32(time.Millisecond) * float32(track_length*float32(BUFFERSIZE)/float32(len(content))) * PRODUCTFACTOR)
			// log.Printf("Time: %v, Track Length: %v, Buffer Size: %v, Content Length: %v, time.Millisecond * DELAY: %v", timer, track_length, BUFFERSIZE, len(content), time.Millisecond*DELAY)
			ticker = time.NewTicker(timer)
		}

		for range ticker.C {
			_, err := tempfile.Read(buffer)

			if err == io.EOF {
				ticker.Stop()
				break
			}
			connectionPool.Broadcast(buffer)
		}
	}
}

func MakeConnection() *Connection {
	return &Connection{bufferChannel: make(chan []byte), buffer: make([]byte, BUFFERSIZE)}
}

func GetConnectionBuffers(conn *Connection) (chan []byte, []byte) {
	return conn.bufferChannel, conn.buffer
}

func PlayAudiofile(connPool *ConnectionPool, filetype string, c *gin.Context) {
	w := c.Writer

	r := c.Request

	w.Header().Add("Content-Type", "audio/"+filetype)
	w.Header().Add("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)

	if !ok {
		log.Println("Could not create flusher")
	}

	connection := MakeConnection()
	connPool.AddConnection(connection)
	log.Printf("%s has connected to the audio stream\n", r.Host)

	for {
		// Receive data from the buffer channel
		bufferChannel, buffer := GetConnectionBuffers(connection)
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
