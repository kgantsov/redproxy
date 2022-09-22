package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/proxy"
	server "github.com/kgantsov/redproxy/pkg/server"
)

func main() {
	level, _ := log.ParseLevel("DEBUG")
	log.SetLevel(level)

	redisPort := flag.Int("redis_port", 46379, "Redis Port")

	flag.Parse()

	proxy := proxy.NewRedisProxy(
		[]*redis.Options{
			{
				Addr:     "localhost:16379",
				Password: "",
				DB:       0,
			},
			{
				Addr:     "localhost:26379",
				Password: "",
				DB:       0,
			},
			{
				Addr:     "localhost:36379",
				Password: "",
				DB:       0,
			},
		},
	)

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
