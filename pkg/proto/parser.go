package proto

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	reader *bufio.Reader
}

type Command struct {
	Name string
	Args []string
}

func NewParser(reader *bufio.Reader) *Parser {
	p := &Parser{}
	p.reader = reader

	return p
}

func (p *Parser) ParseCommand() (*Command, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	if line[0] != '*' {
		return &Command{Name: line}, nil
	}

	argcStr := line[1:]
	argc, err := strconv.ParseUint(argcStr, 10, 64)

	if err != nil || argc < 1 {
		return nil, fmt.Errorf("error parsing number of arguments %s", err)
	}

	args := make([]string, 0, argc)
	for i := 0; i < int(argc); i++ {
		line, err := p.readLine()
		if err != nil {
			return nil, err
		}

		if line[0] != '$' {
			return nil, fmt.Errorf("error parsing argument %s", line)
		}

		argLenStr := line[1:]
		argLen, err := strconv.ParseUint(argLenStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("rror parsing argument length %s", argLenStr)
		}

		arg := make([]byte, argLen+2)
		if _, err := io.ReadFull(p.reader, arg); err != nil {
			return nil, err
		}

		args = append(args, string(arg[0:len(arg)-2]))
	}

	return &Command{Name: strings.ToUpper(args[0]), Args: args[1:]}, nil
}

func (p *Parser) readLine() (string, error) {
	str, err := p.reader.ReadString('\n')
	if err == nil {
		return str[:len(str)-2], err
	}
	return str, err
}
