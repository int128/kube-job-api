package main

import (
	"log"

	"github.com/int128/kube-job-server/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
