package proto

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
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
	log.Info().Msgf("Listening on port: %d", srv.Port)
	defer srv.wg.Done()

	for {
		conn, err := srv.TCPListener.Accept()
		if err != nil {
			select {
			case <-srv.quit:
				return
			default:
				log.Error().Msgf("Fatal error: %s", err.Error())
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
				log.Debug().Msg("Client has been disconnected")
			} else {
				log.Error().Msgf("Error handling request: %v", err)
			}
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal().Msgf("Fatal error: %s", err.Error())
	}
}
