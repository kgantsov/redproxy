package proto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kgantsov/redproxy/pkg/client"

	log "github.com/sirupsen/logrus"
)

type Proto struct {
	parser    *Parser
	responser *Responser
	redis     client.Client
}

func NewProto(redis client.Client, reader io.Reader, writer io.Writer) *Proto {
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
		val, err := p.redis.Get(ctx, cmd.Args[0])
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr(val)
		}
	case "SET":
		err := p.redis.Set(ctx, cmd.Args[0], cmd.Args[1], time.Duration(0))
		if err != nil {
			p.responser.SendStr("")
		} else {
			p.responser.SendStr("OK")
		}
	case "DEL":
		res := p.redis.Del(ctx, cmd.Args...)
		p.responser.SendInt(res)
	case "PING":
		p.responser.SendPong()
	default:
		p.responser.SendError(fmt.Errorf("unknown command '%s'", cmd.Args))
	}
}
