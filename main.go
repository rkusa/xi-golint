package main

import "log"

func main() {
	plugin := NewPlugin()
	plugin.run()

	log.Println("Closing ...")
}
