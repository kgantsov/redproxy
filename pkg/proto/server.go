package proto

import (
	"fmt"
	"io"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	TCPListener *net.TCPListener
	quit        chan interface{}
	redis       *RedisProxy
	Port        int
	wg          sync.WaitGroup
}

func NewServer(redis *RedisProxy, port int) *Server {
	server := &Server{
		redis: redis,
		Port:  port,
		quit:  make(chan interface{}),
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", server.Port))
	checkError(err)
	server.TCPListener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	return server
}

func (srv *Server) ListenAndServe() {
	log.Info("Listening on port: ", srv.Port)
	defer srv.wg.Done()

	for {
		conn, err := srv.TCPListener.Accept()
		if err != nil {
			select {
			case <-srv.quit:
				return
			default:
				log.Error("Fatal error: ", err.Error())
				continue
			}
		}

		srv.wg.Add(1)

		go func() {
			srv.handleClient(conn)
			srv.wg.Done()
		}()
	}
}

func (srv *Server) Stop() {
	close(srv.quit)
	srv.TCPListener.Close()
	srv.wg.Wait()
}

func (srv *Server) handleClient(conn io.ReadWriteCloser) {
	redisProto := NewProto(srv.redis, conn, conn)
	defer conn.Close()

	for {
		err := redisProto.HandleRequest()
		if err != nil {
			if err == io.EOF {
				log.Debug("Client has been disconnected")
			} else {
				log.Error("Error handling request: ", err)
			}
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
}
