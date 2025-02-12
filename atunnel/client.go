package atunnel

import (
	"io"
	"log"
	"net"

	"github.com/Aurascape/smux"
)

// client inner session needs to implement io.ReadWriteCloser
// and is better to implement io.WriteTo() for better performance

type ClientOuterSession struct {
	session *smux.Session
}

func NewClientOuterSession(server string) (*ClientOuterSession, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println("[CLT] failed to connect to server", err)
		return nil, err
	}

	log.Println("[CLT] connected to server", server)

	client, err := smux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &ClientOuterSession{session: client}, nil
}

func (s *ClientOuterSession) Close() {
	s.session.Close()
}

func (s *ClientOuterSession) StartInnerSession(client io.ReadWriteCloser, appName, dstAddr string) {
	meta := make(map[string]interface{})
	meta["app"] = appName
	meta["dst"] = dstAddr
	stream, err := s.session.OpenStream1(&meta)
	if err != nil {
		log.Println("[CLT] failed to open stream", err)
		return
	}

	Pipe(stream, client, 1)
}
