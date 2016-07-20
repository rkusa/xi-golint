package main

import "log"

func main() {
	p := &LintPlugin{}
	err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}
