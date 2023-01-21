package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

// WriteFileToGCS - Writes file to GCS
func WriteFileToGCS(bucket string, obj string, content []byte) (e error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to GCS : %v", err)
	}

	wc := client.Bucket(bucket).Object(obj).NewWriter(ctx)
	wc.ContentType = "text/csv"
	if _, err := wc.Write(content); err != nil {
		return fmt.Errorf("unable to write object %s to bucket %s : %v", obj, bucket, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("createFile: unable to close bucket %s, object %s: %v", bucket, obj, err)

	}

	return nil
}
