package server

import (
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/proto"
	"github.com/kgantsov/redproxy/pkg/proxy"
)

type Server struct {
	Port        int
	TCPListener *net.TCPListener
	redis       proxy.Proxy
}

func NewServer(redis proxy.Proxy, port int) *Server {
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

func (srv *Server) handleClient(redis proxy.Proxy, conn io.ReadWriteCloser) {
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
