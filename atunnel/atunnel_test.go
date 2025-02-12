package atunnel_test

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/Aurascape/smux/atunnel"
)

type atunInnerClient struct {
	id  int
	cnt int
}

func (c *atunInnerClient) Write(data []byte) (n int, err error) {
	log.Printf("[INS] %d rcv from remote: %s", c.id, data)
	return len(data), nil
}

func (c *atunInnerClient) Read(data []byte) (n int, err error) {
	if c.cnt == 0 {
		c.cnt++
		log.Printf("[INS] %d sending Hello", c.id)
		d := []byte("Hello\n")
		return copy(data, d), nil
	}
	time.Sleep(5 * time.Second)
	return 0, io.EOF
}

func (c *atunInnerClient) Close() error {
	log.Printf("[INS] %d closed", c.id)
	return nil
}

func TestMultipleInnerSessions(t *testing.T) {
	listenOn := "127.0.0.1:5253"
	server, err := atunnel.NewATunnelServer(listenOn)
	if err != nil {
		log.Println("failed to listen on", listenOn)
		return
	}

	go server.Run()

	client, _ := atunnel.NewClientOuterSession(listenOn)

	go client.StartInnerSession(&atunInnerClient{id: 1}, "TestApp", "1.1.1.1:80")
	go client.StartInnerSession(&atunInnerClient{id: 2}, "TestApp", "127.0.0.1:1249")
	go client.StartInnerSession(&atunInnerClient{id: 3}, "TestApp", "127.0.0.1:1250")

	time.Sleep(20 * time.Second)
}
