package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/golang/lint"
)

const concurrentRequests = 10

type LintPlugin struct {
	remainingLines int
	receivingLines int
	lines          []string
}

// Implement the io.Reader interface
// func (p *Plugin) Read(p []byte) (n int, err error) {

// }

func (p *LintPlugin) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// log.Println(scanner.Text())
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			return err
		}

		switch msg.Method {
		case MethodNone:
			var res Response
			if err := json.Unmarshal(line, &res); err != nil {
				return err
			}

			// log.Println("Response", res)
			err := p.handleResponse(&res)
			if err != nil {
				return err
			}
		case MethodPing:
			log.Println("ping")
		case MethodPingFromEditor:
			log.Println("ping_from_editor", msg.Params)
			err := p.send(&Request{-1, MethodNLines, nil}) // []struct{}{}})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *LintPlugin) send(req *Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	data = append(data, []byte("\n")...)
	os.Stdout.Write(data)

	return nil
}

func (p *LintPlugin) handleResponse(res *Response) error {
	if res.ID == -1 {
		if f, ok := res.Result.(float64); ok {
			n := int(f)
			p.remainingLines = n
			p.receivingLines = n
			p.lines = make([]string, n)
			for i := 0; i < concurrentRequests; i++ {
				l := p.remainingLines - 1
				if l < 0 {
					break
				}
				err := p.send(&Request{l, MethodGetLine, map[string]int{"line": l}})
				if err != nil {
					return err
				}
				p.remainingLines--
			}
		} else {
			return fmt.Errorf("Unexpected result type, expected an integer, got %v", reflect.TypeOf(res.Result))
		}
	} else {
		// receive line
		i := res.ID
		if i < 0 || i >= len(p.lines) {
			return fmt.Errorf("Received line is out of index, got %v", i)
		}

		if line, ok := res.Result.(string); ok {
			p.lines[i] = line
			if p.remainingLines > 0 {
				l := p.remainingLines - 1
				p.send(&Request{l, MethodGetLine, map[string]int{"line": l}})
				p.remainingLines--
			}

			p.receivingLines--
			if p.receivingLines == 0 {
				// log.Println("Received all lines!")
				if err := p.lint(); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("Unexpected result type, expected a string, got %v", reflect.TypeOf(res.Result))
		}
	}

	return nil
}

func (p *LintPlugin) lint() error {
	src := []byte(strings.Join(p.lines, ""))

	linter := lint.Linter{}
	problems, err := linter.Lint("", src)
	positions := []token.Position{}
	if err != nil {
		if errors, ok := err.(scanner.ErrorList); ok {
			for _, err := range errors {
				log.Println("Lint Error:", err)
				positions = append(positions, err.Pos)
			}
		} else {
			return err
		}
	}

	for _, problem := range problems {
		log.Println("Lint Problem:", problem)
		positions = append(positions, problem.Position)
	}

	for _, pos := range positions {
		start, end := pos.Column-1, len(p.lines[pos.Line-1])
		if start == end {
			start = 0
		}
		// TODO: multiple lint errors per line possible?
		err := p.send(&Request{0, MethodSetLineFgSpans, SetLineFgSpansArgs{
			Line: pos.Line - 1,
			Spans: []Span{
				// TODO: check for out of index errors
				// 4290772992 = 0xFFDB231B
				Span{Start: start, End: end, Fg: 4292551451},
			},
		}})
		if err != nil {
			return err
		}
	}

	return nil
}
