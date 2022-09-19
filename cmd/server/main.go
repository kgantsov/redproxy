package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/proxy"
	server "github.com/kgantsov/redproxy/pkg/server"
)

func main() {
	redisPort := flag.Int("redis_port", 46379, "Redis Port")

	flag.Parse()

	proxy := proxy.NewRedisProxy(
		"localhost",
		"6379",
		"",
		0,
	)

	// client := client.NewMockClient(
	// 	map[string]string{"k1": "v1", "k2": "2", "k3": "value", "year": "2022"},
	// )

	srv := server.NewServer(proxy, *redisPort)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Info(sig)

		log.Info("Stopping the application")

		os.Exit(0)
	}()

	srv.ListenAndServe()
}
