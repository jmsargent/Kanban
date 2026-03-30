package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if len(os.Args) > 2 && os.Args[1] == "--port" {
		port = os.Args[2]
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	fmt.Fprintf(os.Stderr, "kanban-web listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
