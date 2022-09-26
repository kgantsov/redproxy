package proto

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type Responser struct {
	conn io.Writer
}

func NewResponser(conn io.Writer) *Responser {
	r := &Responser{conn}

	return r
}

func (r *Responser) SendError(val error) {
	_, err := fmt.Fprintf(r.conn, "-ERR %s\r\n", val)

	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}
}

func (r *Responser) SendPong() {
	_, err := fmt.Fprintf(r.conn, "+PONG\r\n")
	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}
}

func (r *Responser) SendInt(value int64) {
	_, err := fmt.Fprintf(r.conn, ":%d\r\n", value)

	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}
}

func (r *Responser) SendStr(value string) {
	_, err := fmt.Fprintf(r.conn, "+%s\r\n", value)

	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}
}

func (r *Responser) SendNull() {
	_, err := fmt.Fprintf(r.conn, "$-1\r\n")

	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}
}

func (r *Responser) SendArr(values []string) {
	_, err := fmt.Fprintf(r.conn, "*%d\r\n", len(values))

	if err != nil {
		log.Errorf("Cound not send a aresponse")
	}

	for _, value := range values {
		_, err = fmt.Fprintf(r.conn, "$%d\r\n", len(value))

		if err != nil {
			log.Errorf("Cound not send a aresponse")
		}

		_, err = fmt.Fprintf(r.conn, "%s\r\n", value)

		if err != nil {
			log.Errorf("Cound not send a aresponse")
		}
	}
}
