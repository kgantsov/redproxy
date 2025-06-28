package proto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Proto struct {
	parser    *Parser
	responser *Responser
	redis     *RedisProxy
}

func NewProto(redis *RedisProxy, reader io.Reader, writer io.Writer) *Proto {
	r := bufio.NewReader(reader)
	parser := NewParser(r)
	responser := NewResponser(writer)

	p := &Proto{
		parser:    parser,
		responser: responser,
		redis:     redis,
	}

	return p
}

func (p *Proto) HandleRequest() error {
	var ctx = context.Background()

	cmd, err := p.parser.ParseCommand()
	log.Info().Msgf("GOT A COMMAND %+v", cmd)

	if err != nil {
		if err == io.EOF {
			log.Debug().Msg("Client has been disconnected")
		} else {
			p.responser.SendError(err)
		}

		return err
	}

	log.Info().Msgf("Running '%s' command with args: %+v", cmd.Name, cmd.Args)

	switch cmd.Name {
	case "HELLO":
		p.responser.SendArr([]string{})
	case "GET":
		val, err := p.redis.Get(ctx, cmd.Args[0]).Result()
		if err != nil {
			p.responser.SendNull()
		} else {
			p.responser.SendStr(val)
		}
	case "SET":
		expiration := 0
		multiplier := time.Second

		if len(cmd.Args) == 4 {
			if strings.ToUpper(cmd.Args[2]) == "PX" {
				multiplier = time.Millisecond
			} else if strings.ToUpper(cmd.Args[2]) != "EX" {
				p.responser.SendError(fmt.Errorf("invalid option '%s' for SET command", cmd.Args[2]))
				return nil
			}

			expiration, err = strconv.Atoi(cmd.Args[3])
			if err != nil {
				p.responser.SendError(err)
				return nil
			}

			log.Info().Msgf("Setting expiration to %s %d seconds", cmd.Args[2], expiration)
		}

		err = p.redis.Set(ctx, cmd.Args[0], cmd.Args[1], time.Duration(expiration)*multiplier).Err()
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr("OK")
		}
	case "HGET":
		val, err := p.redis.HGet(ctx, cmd.Args[0], cmd.Args[1]).Result()
		if err != nil {
			p.responser.SendNull()
		} else {
			p.responser.SendStr(val)
		}
	case "HSET":
		err := p.redis.HSet(ctx, cmd.Args[0], cmd.Args[1], cmd.Args[2]).Err()
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr("OK")
		}
	case "DEL":
		res := p.redis.Del(ctx, cmd.Args...).Val()
		p.responser.SendInt(res)
	case "KEYS":
		values := p.redis.Keys(ctx, cmd.Args[0]).Val()
		p.responser.SendArr(values)
	case "APPEND":
		value := p.redis.Append(ctx, cmd.Args[0], cmd.Args[1]).Val()
		p.responser.SendInt(value)
	case "INCR":
		value := p.redis.IncrBy(ctx, cmd.Args[0], 1).Val()
		p.responser.SendInt(value)
	case "INCRBY":
		incrBy, err := strconv.Atoi(cmd.Args[1])

		if err != nil {
			p.responser.SendError(err)
			return nil
		}

		value := p.redis.IncrBy(ctx, cmd.Args[0], int64(incrBy)).Val()
		p.responser.SendInt(value)
	case "DECR":
		value := p.redis.DecrBy(ctx, cmd.Args[0], 1).Val()
		p.responser.SendInt(value)
	case "DECRBY":
		decrBy, err := strconv.Atoi(cmd.Args[1])

		if err != nil {
			p.responser.SendError(err)
			return nil
		}

		value := p.redis.DecrBy(ctx, cmd.Args[0], int64(decrBy)).Val()
		p.responser.SendInt(value)
	case "EXISTS":
		exists := p.redis.Exists(ctx, cmd.Args...).Val()
		if exists > 0 {
			p.responser.SendInt(exists)
		} else {
			p.responser.SendInt(0)
		}
	case "TTL":
		ttl := p.redis.TTL(ctx, cmd.Args[0]).Val()
		p.responser.SendInt(int64(ttl.Seconds()))
	case "EXPIRE":
		expiration, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			p.responser.SendError(err)
			return nil
		}

		res := p.redis.Expire(ctx, cmd.Args[0], time.Duration(expiration)*time.Second)

		if res.Val() {
			p.responser.SendInt(1)
		} else {
			p.responser.SendInt(0)
		}

	case "PING":
		p.responser.SendPong()
	default:
		p.responser.SendError(fmt.Errorf("unsupported command '%s'", cmd.Name))
	}

	return nil
}
