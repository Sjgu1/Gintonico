package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kabukky/httpscerts"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")

}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
	http.Redirect(w, r, "https://127.0.0.1:8081"+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	// Check if the cert files are available.
	err := httpscerts.Check("cert.pem", "key.pem")
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:8081")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
	http.HandleFunc("/", handler)

	// Start the HTTPS server in a goroutine
	go http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", nil)
	// Start the HTTP server and redirect all incoming connections to HTTPS
	err2 := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS))
	log.Fatal(err2)
}