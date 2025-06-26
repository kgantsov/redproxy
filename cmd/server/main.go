package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/kgantsov/redproxy/pkg/proto"
)

var (
	logLevel string
	hostsStr string
	port     int
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano

	flag.StringVar(&logLevel, "log_level", "debug", "Log level")
	flag.StringVar(&hostsStr, "hosts", "localhost:6379,localhost:6380,localhost:6381", "Redis hosts")
	flag.IntVar(&port, "port", 46379, "Redis Port")
	flag.Parse()

	hosts := strings.Split(hostsStr, ",")

	logLevel, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}

	flag.Parse()

	redises := map[string]proto.RedisClient{}

	for _, host := range hosts {
		log.Info().Msgf("Connecting to Redis at %s", host)
		client := redis.NewClient(&redis.Options{
			Addr:     host,
			Password: "",
			DB:       0,
		})
		redises[host] = client
	}

	proxy := proto.NewRedisProxy(redises)

	srv := proto.NewServer(proxy, port)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Info().Msgf("Received signal: %s", sig)

		log.Info().Msg("Stopping the application")

		os.Exit(0)
	}()

	srv.ListenAndServe()
}
