package proto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kgantsov/redproxy/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

type Proto struct {
	parser    *Parser
	responser *Responser
	redis     proxy.Proxy
}

func NewProto(redis proxy.Proxy, reader io.Reader, writer io.Writer) *Proto {
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

		val, err := p.redis.Get(ctx, cmd.Args[0])
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr(val)
		}
		log.Infof("------> ::: %s %s", err, val)
	case "SET":
		log.Infof("=====> SET %+v", cmd.Args)

		err := p.redis.Set(ctx, cmd.Args[0], cmd.Args[1], time.Duration(0))
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr("OK")
		}
	case "DEL":
		log.Infof("=====> DEL %+v", cmd.Args)

		res := p.redis.Del(ctx, cmd.Args...)
		p.responser.SendInt(res)
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
