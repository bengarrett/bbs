//go:build ignore

package main

import (
	"log"
	"net/http"
)

const port = "8070"

func main() {
	fs := http.FileServer(http.Dir("web"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fs.ServeHTTP(w, r)
	})
	log.Println("Serving web/ at http://localhost:" + port + " (Ctrl+C to stop) ...")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
