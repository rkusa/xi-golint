package main

import (
	"fmt"
	"log"
	"time"

	xi "github.com/rkusa/xi-peer"
)

var done chan bool
var peer *xi.Peer

func main() {
	done = make(chan bool)
	peer = xi.New()
	peer.Handle("ping", handlePing)
	peer.Handle("ping_from_editor", handlePingFromEditor)

	<-done
	log.Println("Closing ...")
}

func handlePing(params interface{}) {
	log.Println("ping received")
}

func handlePingFromEditor(params interface{}) {
	log.Println("ping_from_editor received")
	retrieveAllLines(10)
}

func retrieveAllLines(concurrency int) {
	var n float64 = 0
	if err := peer.CallSync("n_lines", nil, &n); err != nil {
		log.Fatal(err)
	}

	log.Println("n_lines", n)

	start := time.Now()

	response := make(chan *xi.Call, concurrency)
	lines := make([]string, int(n))
	remaining := len(lines)
	receiving := 0

	for remaining > 0 || receiving > 0 {
		if receiving >= concurrency || remaining == 0 {
			// wait for a response to arrive, before making a new request
			call := <-response
			if call.Error != nil {
				log.Fatal(call.Error)
			}
			receiving--
		}

		if remaining == 0 {
			break
		}

		lnr := remaining - 1
		peer.Call("get_line", map[string]int{"line": lnr}, &lines[lnr], response)

		remaining--
		receiving++
	}

	elapsed := time.Since(start)
	alert := fmt.Sprintf("Retrieving all %d lines took %s (with %d concurrent requests)", len(lines), elapsed, concurrency)
	log.Println(alert)

	if err := peer.CallSync("alert", map[string]string{"msg": alert}, nil); err != nil {
		log.Fatal(err)
	}

	done <- true

	// if err := p.lint(); err != nil {
	// 	return err
	// }
}
