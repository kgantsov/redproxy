package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/proto"
)

func main() {
	level, _ := log.ParseLevel("DEBUG")
	log.SetLevel(level)

	redisPort := flag.Int("redis_port", 46379, "Redis Port")

	flag.Parse()

	redises := map[string]proto.RedisClient{}

	redisesOptions := []*redis.Options{
		{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		{
			Addr:     "localhost:6380",
			Password: "",
			DB:       0,
		},
		{
			Addr:     "localhost:6381",
			Password: "",
			DB:       0,
		},
	}

	for _, redisOptions := range redisesOptions {
		client := redis.NewClient(redisOptions)
		redises[redisOptions.Addr] = client
	}

	proxy := proto.NewRedisProxy(redises)

	srv := proto.NewServer(proxy, *redisPort)

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
