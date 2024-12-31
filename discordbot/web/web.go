package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func Setup(ctx context.Context) {
	// Service static file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// hello world
		fmt.Fprintf(w, "Hello, world!")
	})
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := 8080
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port, _ = strconv.Atoi(envPort)
	}

	// Listen
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	fmt.Printf("Web server running on port %d\n", port)
}
