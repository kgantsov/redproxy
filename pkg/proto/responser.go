package proto

import (
	"fmt"
	"io"
)

type Responser struct {
	conn io.Writer
}

func NewResponser(conn io.Writer) *Responser {
	r := &Responser{conn}

	return r
}

func (r *Responser) SendError(err error) {
	r.conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err)))
}

func (r *Responser) SendPong() {
	r.conn.Write([]byte("+PONG\r\n"))
}

func (r *Responser) SendInt(value int64) {
	r.conn.Write([]byte(fmt.Sprintf(":%d\r\n", value)))
}

func (r *Responser) SendStr(value string) {
	r.conn.Write([]byte(fmt.Sprintf("+%s\r\n", value)))
}

func (r *Responser) SendArr(values []string) {
	r.conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(values))))
	for _, value := range values {
		r.conn.Write([]byte(fmt.Sprintf("$%d\r\n", len(value))))
		r.conn.Write([]byte(fmt.Sprintf("%s\r\n", value)))
	}
}
