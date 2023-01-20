package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// http handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HttpHandler Handles the HTTP call
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	reqHeadersBytes, err := json.Marshal(r.Header)
	if err != nil {
		log.Println("Could not Marshal Req Headers")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBodyBytes, err := json.Marshal(r.Body)
	if err != nil {
		log.Println("Could not Marshal Req Body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Set Response Code
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}

	log.Printf("Request Headers: %s", reqHeadersBytes)
	log.Printf("Request Body: %s", reqBodyBytes)

}
