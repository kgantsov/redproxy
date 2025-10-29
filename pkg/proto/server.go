package proto

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog/log"
)

type Server struct {
	TCPListener *net.TCPListener
	quit        chan any
	redis       *RedisProxy
	Port        int
	wg          sync.WaitGroup

	router  *fiber.App
	Metrics *PrometheusMetrics
}

func NewServer(redis *RedisProxy, port int) *Server {
	router := fiber.New()

	registry := prometheus.NewRegistry()

	prom := fiberprometheus.NewWithRegistry(
		registry, "redproxy", "redproxy", "redproxy", map[string]string{},
	)
	prom.RegisterAt(router, "/metrics")
	router.Use(prom.Middleware)
	registry.Register(collectors.NewGoCollector())

	server := &Server{
		redis:   redis,
		Port:    port,
		quit:    make(chan interface{}),
		Metrics: NewPrometheusMetrics(registry, "redproxy", "redproxy"),
		router:  router,
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", server.Port))
	checkError(err)
	server.TCPListener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	return server
}

func (srv *Server) ListenAndServe() {
	go func() {
		err := srv.router.Listen(":9090")
		if err != nil {
			log.Error().Msgf("Fatal error: %s", err.Error())
		}
	}()

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
		srv.Metrics.Connections.With(prometheus.Labels{}).Inc()

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
	redisProto := NewProto(srv.Metrics, srv.redis, conn, conn)
	defer conn.Close()

	for {
		err := redisProto.HandleRequest()
		if err != nil {
			if err == io.EOF {
				log.Debug().Msg("Client has been disconnected")
				srv.Metrics.Connections.With(prometheus.Labels{}).Dec()
			} else {
				log.Error().Msgf("Error handling request: %v", err)
			}
			return
		}
		srv.Metrics.CommandsProxiedTotal.With(prometheus.Labels{}).Inc()
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal().Msgf("Fatal error: %s", err.Error())
	}
}
