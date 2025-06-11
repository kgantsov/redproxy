package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/proto"
)

var (
	logLevel string
	hostsStr string
	port     int
)

func main() {
	flag.StringVar(&logLevel, "log_level", "debug", "Log level")
	flag.StringVar(&hostsStr, "hosts", "localhost:6379,localhost:6380,localhost:6381", "Redis hosts")
	flag.IntVar(&port, "port", 46379, "Redis Port")
	flag.Parse()

	hosts := strings.Split(hostsStr, ",")

	level, _ := log.ParseLevel(logLevel)
	log.SetLevel(level)

	flag.Parse()

	redises := map[string]proto.RedisClient{}

	for _, host := range hosts {
		fmt.Printf("Connecting to Redis at %s\n", host)
		client := redis.NewClient(&redis.Options{
			Addr:     host,
			Password: "",
			DB:       0,
		})
		redises[host] = client
		// client.MSet()
	}

	proxy := proto.NewRedisProxy(redises)

	srv := proto.NewServer(proxy, port)

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
