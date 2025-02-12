package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Aurascape/smux/atunnel"
)

type aHttpClient struct {
	cnt int
}

func (c *aHttpClient) Write(data []byte) (n int, err error) {
	log.Printf("[http] recv: %s", data)
	return len(data), nil
}

func (c *aHttpClient) Read(data []byte) (n int, err error) {
	if c.cnt == 0 {
		c.cnt++
		request := []byte("GET / HTTP/1.1\r\nHost: 1.1.1.1\r\nUser-Agent: atunnel/0.1\r\nAccept: */*\r\n\r\n")
		i := copy(data, request)
		log.Printf("[http] sending GET request:\n%s", string(request))
		return i, nil
	}
	return 0, io.EOF
}

// this one will be called as it has better performance
func (c *aHttpClient) WriteTo(w io.Writer) (n int64, err error) {
	request := []byte("GET / HTTP/1.1\r\nHost: 1.1.1.1\r\nUser-Agent: atunnel/0.1\r\nAccept: */*\r\n\r\n")
	log.Printf("[http] sending request:\n%s", string(request))

	for {
		var i int
		i, err = w.Write(request)
		n += int64(i)
		if err != nil || i == len(request) {
			break
		}
		request = request[i:]
	}

	return n, err
}

func (c *aHttpClient) Close() error {
	log.Println("[http] closed")
	return nil
}

// "-s addr:port" to start it as a server
// "-c addr:port" to start it as a client, a stardard http GET request will be sent to the server
// "-d" to enable debug log
// "-t" to set destination inner session address
func main() {
	var serverAddr string
	var connectAddr string
	var destAddr string
	flag.StringVar(&serverAddr, "s", "", "server address")
	flag.StringVar(&connectAddr, "c", "", "connect to address")
	flag.StringVar(&destAddr, "t", "", "destination inner session address")
	flag.Parse()

	if serverAddr != "" {
		server, err := atunnel.NewATunnelServer(serverAddr)
		if err == nil {
			server.Run()
		} else {
			log.Fatal(err)
		}
	} else if connectAddr != "" {
		client, err := atunnel.NewClientOuterSession(connectAddr)
		if err != nil {
			log.Fatal(err)
		}

		go client.Run()

		if destAddr == "" {
			destAddr = "127.0.0.1:80"
		}

		for i := 0; i < 10; i++ {
			go client.StartInnerSession(&aHttpClient{}, "ATunnelClient", destAddr)
			time.Sleep(1 * time.Second)
		}
	} else {
		fmt.Println("Usage: atunnel -s addr:port | -c addr:port")
	}
}
