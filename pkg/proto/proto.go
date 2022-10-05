package proto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
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

func (p *Proto) HandleRequest() {
	var ctx = context.Background()

	cmd, err := p.parser.ParseCommand()
	log.Infof("GOT A COMMAND %+v", cmd)

	if err != nil {
		if err == io.EOF {
			log.Debug("Client has been disconnected")
		} else {
			p.responser.SendError(err)
		}

		return
	}

	switch cmd.Name {
	case "HELLO":
		p.responser.SendArr([]string{})
	case "GET":
		log.Infof("=====> GET %+v", cmd.Args)

		val, err := p.redis.Get(ctx, cmd.Args[0]).Result()
		if err != nil {
			p.responser.SendNull()
		} else {
			p.responser.SendStr(val)
		}

		log.Infof("------> ::: %s %s", err, val)
	case "SET":
		log.Infof("=====> SET %+v", cmd.Args)

		err := p.redis.Set(ctx, cmd.Args[0], cmd.Args[1], time.Duration(0)).Err()
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr("OK")
		}
	case "DEL":
		log.Infof("=====> DEL %+v", cmd.Args)

		res := p.redis.Del(ctx, cmd.Args...).Val()
		p.responser.SendInt(res)
	case "KEYS":
		log.Infof("=====> KEYS %+v", cmd.Args)

		values := p.redis.Keys(ctx, cmd.Args[0]).Val()
		p.responser.SendArr(values)
	case "APPEND":
		log.Infof("=====> APPEND %+v", cmd.Args)

		value := p.redis.Append(ctx, cmd.Args[0], cmd.Args[1]).Val()
		p.responser.SendInt(value)
	case "INCR":
		log.Infof("=====> INCR %+v", cmd.Args)

		value := p.redis.IncrBy(ctx, cmd.Args[0], 1).Val()
		p.responser.SendInt(value)
	case "INCRBY":
		log.Infof("=====> INCRBY %+v", cmd.Args)

		decrBy, err := strconv.Atoi(cmd.Args[1])

		if err != nil {
			p.responser.SendError(err)
			return
		}

		value := p.redis.IncrBy(ctx, cmd.Args[0], int64(decrBy)).Val()
		p.responser.SendInt(value)
	case "DECR":
		log.Infof("=====> DECR %+v", cmd.Args)

		value := p.redis.DecrBy(ctx, cmd.Args[0], 1).Val()
		p.responser.SendInt(value)
	case "DECRBY":
		log.Infof("=====> DECRBY %+v", cmd.Args)

		decrBy, err := strconv.Atoi(cmd.Args[1])

		if err != nil {
			p.responser.SendError(err)
			return
		}

		value := p.redis.DecrBy(ctx, cmd.Args[0], int64(decrBy)).Val()
		p.responser.SendInt(value)
	// case "MGET":
	// 	log.Infof("=====> MGET %+v", cmd.Args)

	// 	var results []string

	// 	for _, key := range cmd.Args {
	// 		value, ok := p.store[key]
	// 		if ok {
	// 			results = append(results, value)
	// 		}
	// 	}

	// 	p.responser.sendArr(results)

	// case "MSET":
	// 	log.Infof("=====> MSET %+v", cmd.Args)

	// 	for i := 0; i < len(cmd.Args); i += 2 {
	// 		key := cmd.Args[i]
	// 		value := cmd.Args[i+1]
	// 		p.store[key] = value
	// 	}

	// 	p.responser.sendStr("OK")

	case "PING":
		p.responser.SendPong()
	default:
		p.responser.SendError(fmt.Errorf("unknown command '%s'", cmd.Args))
	}
}
