package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

const (
	gcsbucket      = "webhook-looker-6814"
	lookerdatafile = "dashboard-svod_vix_daily_kpis/users_2.csv"
)

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
		log.Printf("Could not unmarshal request Body : %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if rB.Attachment.Extension != "zip" && rB.Attachment.Mimetype != "application/zip;base64" {
		log.Printf("Recieved unsupported payload")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DataFromZipToGCS(rB.Attachment.Data, lookerdatafile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// DataFromZipToGCS - gets data from base64 encoded zip file
func DataFromZipToGCS(b64 string, fname string) (e error) {
	rawZip, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("could not decode request Body Data : %v", err)
	}

	// Read rawZip with zip module and loop over contents
	archive, err := zip.NewReader(bytes.NewReader(rawZip), bytes.NewReader(rawZip).Size())
	if err != nil {
		return fmt.Errorf("could not unzip the payload data : %v", err.Error()[0:50])
	}

	for _, file := range archive.File {
		if file.Name == fname {
			// write to GCS
			log.Printf("%s\n", file.Name)
			flatfile, err := file.Open()
			if err != nil {
				return fmt.Errorf("could not retrieve contents of %s : %v ", file.Name, err)
			}

			filearray, _ := io.ReadAll(flatfile)
			WriteFileToGCS(gcsbucket, lookerdatafile, filearray)

			return nil
		}

	}

	return nil
}
