package main

import "fmt"

type ResponseMethod int

const (
	MethodNone ResponseMethod = iota
	MethodPing
	MethodPingFromEditor
)

func (mt *ResponseMethod) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"ping"`:
		*mt = MethodPing
	case `"ping_from_editor"`:
		*mt = MethodPingFromEditor
	default:
		return fmt.Errorf("invalid method %s", b)
	}

	return nil
}

type Message struct {
	Method ResponseMethod
	Params interface{}
}

type Response struct {
	ID     int
	Result interface{}
}

type ResponseType int

const (
	ResponseNLines ResponseType = iota
)

type RequestMethod int

const (
	MethodNLines RequestMethod = iota
	MethodGetLine
	MethodSetLineFgSpans
)

func (rm RequestMethod) MarshalJSON() ([]byte, error) {
	switch rm {
	case MethodNLines:
		return []byte(`"n_lines"`), nil
	case MethodGetLine:
		return []byte(`"get_line"`), nil
	case MethodSetLineFgSpans:
		return []byte(`"set_line_fg_spans"`), nil
	}
	return nil, fmt.Errorf("failed to marshal request method %s to json", rm)
}

type Request struct {
	ID     int           `json:"id"`
	Method RequestMethod `json:"method"`
	Params interface{}   `json:"params"`
}

type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Fg    int `json:"fg"`
}

type SetLineFgSpansArgs struct {
	Line  int    `json:"line"`
	Spans []Span `json:"spans"`
}
