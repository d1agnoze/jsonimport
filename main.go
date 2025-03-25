package main

import (
	"d1agnoze/jsonimport/internal/server"
)

func main() {
	server := server.New()
	server.Run()
}
