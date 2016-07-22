package main

import (
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"
	"unicode"
)

type plugin struct {
	client *Client
	server *rpc.Server
}

func NewPlugin() *plugin {
	p := new(plugin)

	// create an JSON-RPC server (currently only to receive ping and
	// ping_from_editor)
	server := rpc.NewServer()
	if err := server.RegisterName("Plugin", p); err != nil {
		log.Fatal(err)
	}

	p.client = NewClient()
	p.server = server

	return p
}

func (p *plugin) run() {
	// start server and wait for ping_from_editor
	codec := &serverCodec{jsonrpc.NewServerCodec(NewStdinStdoutConn())}
	// codec := jsonrpc.NewServerCodec(&conn{os.Stdin, os.Stdout})

	// When running both client and server on Stdin/Stdout they steel each other
	// the responses. Therefore, for now, ping and ping_from_editor are
	// received manually.
	// p.server.ServeCodec(codec)
	for i := 0; i < 2; i++ {
		if err := p.server.ServeRequest(codec); err != nil {
			log.Fatal(err)
		}
	}
}

func (p *plugin) retrieveAllLines(concurrency int) {
	var n float64 = 0
	if err := p.client.CallSync("n_lines", nil, &n); err != nil {
		log.Fatal(err)
	}

	log.Println("n_lines", n)

	start := time.Now()

	lines := make([]string, int(n))
	remaining := len(lines)
	receiving := 0

	for remaining > 0 {
		if receiving >= concurrency {
			// wait for a response to arrive, before making a new request
			call := <-p.client.Recv
			if call.Error != nil {
				log.Fatal(call.Error)
			}
			receiving--
		}

		lnr := remaining - 1
		if lnr < 0 {
			break
		}

		p.client.Call("get_line", lnr, &lines[lnr])
		remaining--
		receiving++
	}

	elapsed := time.Since(start)
	log.Printf("Retrieving all %d lines took %s (with %d concurrent requests)", len(lines), elapsed, concurrency)

	// if err := p.lint(); err != nil {
	// 	return err
	// }
}

func (p *plugin) Ping(args int, reply *int) error {
	log.Println("ping received")
	return nil
}

func (p *plugin) PingFromEditor(args int, reply *int) error {
	log.Println("ping_from_editor received")
	p.retrieveAllLines(1)
	return nil
}

// func (p *LintPlugin) lint() error {
// 	src := []byte(strings.Join(p.lines, ""))
//
// 	linter := lint.Linter{}
// 	problems, err := linter.Lint("", src)
// 	positions := []token.Position{}
// 	if err != nil {
// 		if errors, ok := err.(scanner.ErrorList); ok {
// 			for _, err := range errors {
// 				log.Println("Lint Error:", err)
// 				positions = append(positions, err.Pos)
// 			}
// 		} else {
// 			return err
// 		}
// 	}
//
// 	for _, problem := range problems {
// 		log.Println("Lint Problem:", problem)
// 		positions = append(positions, problem.Position)
// 	}
//
// 	for _, pos := range positions {
// 		start, end := pos.Column-1, len(p.lines[pos.Line-1])
// 		if start == end {
// 			start = 0
// 		}
// 		// TODO: multiple lint errors per line possible?
// 		err := p.send(&Request{0, MethodSetLineFgSpans, SetLineFgSpansArgs{
// 			Line: pos.Line - 1,
// 			Spans: []Span{
// 				// TODO: check for out of index errors
// 				// 4290772992 = 0xFFDB231B
// 				Span{Start: start, End: end, Fg: 4292551451},
// 			},
// 		}})
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }

// this custom server codec is used for making Go's jsonrpc compatible with
// Xi's JSON-RPC endpoint.
type serverCodec struct {
	rpc.ServerCodec
}

func (sc *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	if err := sc.ServerCodec.ReadRequestHeader(r); err != nil {
		return err
	}

	// In Go only methods of structs can be exported and their name is used
	// as the JSON-RPC method accordinglty. E.g. Plugin.Ping
	// This requires to rewrite Xi's method names;
	// e.g. from ping_from_editor to Plugin.PingFromEditor
	r.ServiceMethod = "Plugin." + snakeCaseToCamelCase(r.ServiceMethod)

	return nil
}

func (sc *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	// ping and ping_from_editor do not expect a response, so don't send one
	return nil
}

func snakeCaseToCamelCase(s string) string {
	words := strings.Split(s, "_")

	for i, w := range words {
		if len(w) == 0 {
			continue
		}

		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}

	return strings.Join(words, "")
}
