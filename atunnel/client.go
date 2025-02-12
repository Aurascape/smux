package atunnel

import (
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/Aurascape/smux"
)

// client inner session needs to implement io.ReadWriteCloser
// and is better to implement io.WriteTo() for better performance

type ClientOuterSession struct {
	config  *smux.Config
	session atomic.Pointer[smux.Session]
	dialTo  string
	closed  atomic.Bool
}

func NewClientOuterSession(server string) (*ClientOuterSession, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println("[CLT] failed to connect to server", err)
		return nil, err
	}

	log.Println("[CLT] connected to server", server)

	config := smux.DefaultConfig()
	config.Version = 2

	client, err := smux.Client(conn, config)
	if err != nil {
		return nil, err
	}

	s := &ClientOuterSession{
		config: config,
		dialTo: server,
	}
	s.session.Store(client)
	s.closed.Store(false)
	return s, nil
}

// this function will keep trying to reconnect to the server
func (s *ClientOuterSession) Run() {
	for !s.closed.Load() {
		client := s.session.Load()

		if client != nil {
			ch := client.CloseChan()
			<-ch
			log.Println("[CLT] session closed", client.LocalAddr())
			s.session.Store(nil)
		}

		newConn, err := net.Dial("tcp", s.dialTo)
		if err != nil {
			log.Println("[CLT] failed to connect to server", err)
			time.Sleep(1 * time.Second)
			continue
		}

		newClient, err := smux.Client(newConn, s.config)
		if err != nil {
			log.Println("[CLT] failed to create smux client", err)
			time.Sleep(1 * time.Second)
			continue
		}

		s.session.Store(newClient)
	}
}

func (s *ClientOuterSession) Close() {
	if s.closed.CompareAndSwap(false, true) {
		session := s.session.Load()
		if session != nil {
			session.Close()
		}
	}
}

func (s *ClientOuterSession) StartInnerSession(client io.ReadWriteCloser, appName, dstAddr string) error {
	meta := make(map[string]interface{})
	meta["app"] = appName
	meta["dst"] = dstAddr

	session := s.session.Load()
	if session == nil {
		return io.ErrClosedPipe
	}

	stream, err := session.OpenStream1(&meta)
	if err != nil {
		log.Println("[CLT] failed to open stream", err)
		session.Close()
		return err
	}

	Pipe(stream, client, 1)
	return nil
}
