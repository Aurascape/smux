package atunnel

import (
	"io"
	"log"
	"net"

	"github.com/Aurascape/smux"
)

type ServerOuterSession struct {
	session *smux.Session
}

type server struct {
	listener net.Listener
}

func newServerOuterSession(conn net.Conn) *ServerOuterSession {
	session, err := smux.Server(conn, nil)
	if err != nil {
		log.Println("[SVR] failed to create smux session", err)
		return nil
	}
	return &ServerOuterSession{session: session}
}

func NewATunnelServer(listenOn string) (*server, error) {
	sock, err := net.Listen("tcp", listenOn)
	if err != nil {
		return nil, err
	}

	return &server{listener: sock}, nil
}

func (s *server) Run() {
	sock := s.listener
	for {
		conn, err := sock.Accept()
		if err != nil {
			break
		}
		go s.handleClient(conn)
	}

	sock.Close()
}

func (s *server) handleClient(conn net.Conn) {
	sos := newServerOuterSession(conn)
	log.Printf("[SVR] serving tunnel client `%s`", conn.RemoteAddr())

	for {
		if stream, err := sos.session.AcceptStream(); err == nil {
			go func(s io.ReadWriteCloser) {
				if meta := stream.Meta(); meta != nil {
					dst := meta["dst"].(string)
					app := meta["app"]

					dstConn, err := net.Dial("tcp", dst)
					if err != nil {
						log.Printf("[SVR] failed to connect to %s: %s", dst, err)
						stream.Close()
						return
					}

					log.Printf("[SVR] %s connected to %s", app, dst)
					Pipe(stream, dstConn, 1)
				} else {
					stream.Close()
					return
				}
			}(stream)
		} else {
			return
		}
	}
}
