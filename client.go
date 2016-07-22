package main

import (
	"bufio"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type Client struct {
	*rpc.Client
	Recv chan *rpc.Call
}

func NewClient() *Client {
	return &Client{
		Client: jsonrpc.NewClient(NewStdinStdoutConn()),
		Recv:   make(chan *rpc.Call, 128),
	}
}

func (c *Client) CallSync(method string, arg interface{}, reply interface{}) error {
	return c.Client.Call(method, arg, reply)
}

func (c *Client) Call(method string, arg interface{}, reply interface{}) {
	c.Client.Go(method, arg, reply, c.Recv)
}

// ReadWriteCloser
type conn struct {
	io.Reader
	io.Writer
}

func NewStdinStdoutConn() *conn {
	return &conn{
		NewLineReader(os.Stdin),
		os.Stdout,
	}
}

func (c *conn) Close() error {
	// TODO: necessary to close Stdin or Stdout?
	return nil
}

type lineReader struct {
	scanner *bufio.Scanner
	stopped bool // TODO: mutex for stopped?
	buf     []byte
}

func NewLineReader(r io.Reader) *lineReader {
	return &lineReader{bufio.NewScanner(r), false, nil}
}

func (lr *lineReader) Read(p []byte) (int, error) {
	if lr.stopped {
		return 0, io.EOF
	}

	var line []byte
	if len(lr.buf) > 0 {
		line = lr.buf
		lr.buf = nil
	} else {
		if cont := lr.scanner.Scan(); !cont {
			lr.stopped = true
			if err := lr.scanner.Err(); err != nil {
				return 0, err
			} else {
				return 0, io.EOF
			}
		}

		line = lr.scanner.Bytes()
		// log.Println("Recv", string(line))
	}

	if len(line) > len(p) {
		lr.buf = line[len(p):]
	}

	return copy(p, line), nil
}
