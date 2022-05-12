package main

import (
	"fmt"
	"os"

	"github.com/p1nant0m/xdp-tracing/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("you should provide the config.yaml path")
	}
	service.Service()
}
