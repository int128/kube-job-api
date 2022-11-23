package main

import (
	"log"

	"github.com/int128/kube-job-server/pkg/manager"
)

func main() {
	if err := manager.Run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
