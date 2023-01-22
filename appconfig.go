package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

// Appconfig holds the application configuration
type AppConfig struct {
	Dashboards []Dashboard `json:"dashboards,omitempty"`
}

type Dashboard struct {
	Name     string `json:"name,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Archives []struct {
		Filename          string `json:"filename"`
		Destinationprefix string `json:"destinationprefix"`
	} `json:"archives"`
}

func (a *AppConfig) InitConfig(obj string) (err error) {
	gcs, err := url.Parse(obj)
	if err != nil {
		return fmt.Errorf("could not parse GCS URL: %s error : %v", obj, err)
	}

	f, err := gcsDownloadFile(gcs.Host, strings.Trim(gcs.Path, "/"))
	if err != nil {
		return fmt.Errorf("could not download the object : %s, error: %v", obj, err)
	}

	err = json.Unmarshal(f, &a)
	if err != nil {
		return fmt.Errorf("could not parse the config file %v", f)
	}

	return nil
}

// https://cloud.google.com/storage/docs/downloading-objects#storage-download-object-go
// gcsdownloadFile downloads an object to a file.
func gcsDownloadFile(bucket, object string) (filebyte []byte, err error) {
	log.Printf("Trying to get configuration object %s from bucket %s", object, bucket)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("could not form storage client %v", err)
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Printf("Object(%q).NewReader: %v", object, err)
		return nil, fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	i, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("could not read the object : %v", err)
		return nil, fmt.Errorf("could not read the object : %v", err)
	}

	return i, nil
}
