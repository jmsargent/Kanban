package main

import (
	"fmt"
	"os"

	"github.com/jmsargent/kanban/internal/adapters/web"
)

func main() {
	addr := ":8080"
	for i := 1; i < len(os.Args)-1; i++ {
		switch os.Args[i] {
		case "--port":
			addr = ":" + os.Args[i+1]
		case "--addr":
			addr = os.Args[i+1]
		}
	}

	server := web.NewServer(addr)
	fmt.Fprintf(os.Stderr, "kanban-web listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
