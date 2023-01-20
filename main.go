package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// https://github.com/GoogleCloudPlatform/golang-samples/blob/HEAD/run/logging-manual/main.go
func init() {
	// Disable log prefixes such as the default timestamp.
	// Prefix text prevents the message from being parsed as JSON.
	// A timestamp is added when shipping logs to Cloud Logging.
	log.SetFlags(0)
}

func main() {
	// init Looker Webhook validation
	validation := lookerWebhook{}
	validation.LookerInstance = os.Getenv("X-Looker-Instance")
	validation.LookerWebhookToken = os.Getenv("X-Looker-Webhook-Token")

	// http handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r, validation)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HttpHandler Handles the HTTP call
func HttpHandler(w http.ResponseWriter, r *http.Request, v lookerWebhook) {
	// validate the Looker headers
	if v.LookerInstance != r.Header.Get("X-Looker-Instance") || v.LookerWebhookToken != r.Header.Get("X-Looker-Webhook-Token") {
		log.Println("Could Not validate Looker headers")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var rB lookerBody

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&rB)
	if err != nil {
		log.Printf("Could not unmarshal request Body : %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if rB.Attachment.Extension != "zip" && rB.Attachment.Mimetype != "application/zip;base64" {
		log.Printf("Recieved unsupported payload")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rawZip, err := base64.StdEncoding.DecodeString(rB.Attachment.Data)
	if err != nil {
		log.Printf("Could not decode request Body Data : %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ioreader := bytes.NewReader(rawZip)

	archive, err := zip.NewReader(ioreader, ioreader.Size())
	if err != nil {
		log.Printf("Could not unzip the payload data : %s", err.Error()[0:50])
		http.Error(w, err.Error()[0:50], http.StatusBadRequest)
		return
	}

	for _, file := range archive.File {
		log.Printf("%s\n", file.Name)

	}

	w.WriteHeader(http.StatusOK)

}
