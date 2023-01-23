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
	"time"
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

	// get Application configuration from EnvVar configfile which is an object in GCS
	var config AppConfig
	err := config.InitConfig(os.Getenv("configfile"))
	if err != nil {
		log.Fatalf("Could not fetch the application config %s, error : %v\n", os.Getenv("configfile"), err)
	}

	log.Printf("AppConfig : %v", config)

	// http handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r, &config, validation)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HttpHandler Handles the HTTP call
func HttpHandler(w http.ResponseWriter, r *http.Request, appconfig *AppConfig, v lookerWebhook) {
	// validate the Looker headers
	log.Printf("Recieved request with X-Looker-Instance: %s, X-Looker-Webhook-Token: %s",
		r.Header.Get("X-Looker-Instance"),
		r.Header.Get("X-Looker-Webhook-Token"))
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
		http.Error(w, "Request Body data not supported", http.StatusBadRequest)
		return
	}

	if rB.Attachment.Extension != "zip" && rB.Attachment.Mimetype != "application/zip;base64" {
		log.Printf("Recieved unsupported payload")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Webhook for Dashboard : %s \n", rB.ScheduledPlan.Title)

	// Check to see if we have a configuration for the request
	found := false
	for i, dashboard := range appconfig.Dashboards {
		// find Configuration for Request Body
		if dashboard.Name == rB.ScheduledPlan.Title {
			found = true
			log.Printf("Found configuration for request dashboard : %s", dashboard.Name)
			err = DataFromZipToGCS(rB.Attachment.Data, &dashboard)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if !found && i == len(appconfig.Dashboards) {
			log.Printf("Could not find configuration for request dashboard : %s", dashboard.Name)
			w.WriteHeader(http.StatusNotImplemented)
		}
	}

	w.WriteHeader(http.StatusOK)

}

// DataFromZipToGCS - gets data from base64 encoded zip file
func DataFromZipToGCS(b64 string, dashboard *Dashboard) (e error) {
	rawZip, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("could not decode request Body Data : %v", err)
	}

	// Read rawZip with zip module and loop over contents
	archive, err := zip.NewReader(bytes.NewReader(rawZip), bytes.NewReader(rawZip).Size())
	if err != nil {
		return fmt.Errorf("could not unzip the payload data : %v", err.Error()[0:50])
	}

	log.Printf("looping throught he archive file ... ")
	for _, file := range archive.File {
		for _, archive := range dashboard.Archives {
			if file.Name == archive.Filename {
				targetobj := archive.Destinationprefix + "-" + fmt.Sprintf("%v", time.Now().Unix()) + ".csv"
				// write to GCS
				log.Printf("Transfering %s to bucket: %s/%s\n", file.Name, dashboard.Bucket, targetobj)
				flatfile, err := file.Open()
				if err != nil {
					return fmt.Errorf("could not retrieve contents of %s : %v ", file.Name, err)
				}

				filearray, _ := io.ReadAll(flatfile)
				err = WriteFileToGCS(dashboard.Bucket, targetobj, filearray)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
