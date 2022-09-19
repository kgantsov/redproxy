package server

import (
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/client"
	"github.com/kgantsov/redproxy/pkg/proto"
)

type Server struct {
	Port        int
	TCPListener *net.TCPListener
	redis       client.Client
}

func NewServer(redis client.Client, port int) *Server {
	server := &Server{redis: redis, Port: port}

	return server
}

func (srv *Server) ListenAndServe() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", srv.Port))
	checkError(err)
	srv.TCPListener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Info("Listening on port: ", srv.Port)

	for {
		conn, err := srv.TCPListener.Accept()
		if err != nil {
			log.Error("Fatal error: ", err.Error())
			continue
		}
		go srv.handleClient(srv.redis, conn)
	}
}

func (srv *Server) handleClient(redis client.Client, conn io.ReadWriteCloser) {
	redisProto := proto.NewProto(redis, conn, conn)
	defer conn.Close()

	for {
		redisProto.HandleRequest()
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
}
