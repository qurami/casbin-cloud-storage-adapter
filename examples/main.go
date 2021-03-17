package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"github.com/casbin/casbin/v2"
	cloudstorageadapter "github.com/qurami/casbin-cloud-storage-adapter/v1"
)

func main() {
	// Initialize a Google Cloud Storage client
	// There are many ways, this is the quickest one.
	// You could need a different one according to your configuration,
	// please see https://pkg.go.dev/cloud.google.com/go/storage
	cloudStorageClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create a new cloudstorageadapter.Adapter
	adapter, err := cloudstorageadapter.NewAdapter(
		cloudStorageClient,
		"myBucketName",
		"path/to/policies.csv",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use the adapter in the casbin.NewEnforcer constructor
	enforcer, err := casbin.NewEnforcer("rbac_model.conf", adapter)
	if err != nil {
		log.Fatal(err)
	}

	// Use the enforcer as usual
	roles, err := enforcer.GetImplicitRolesForUser("alice")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(roles)
}
