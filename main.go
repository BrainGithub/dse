package main

import (
    "dse/internal/fileserver"
    "log"
)

func main() {
	log.Print("dse main starting.")
	fileserver.StartApp()
}
