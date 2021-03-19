# Casbin Cloud Storage Adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/qurami/casbin-cloud-storage-adapter)](https://goreportcard.com/report/github.com/qurami/casbin-cloud-storage-adapter)
[![Build Status](https://travis-ci.com/casbin/casbin.svg?branch=master)](https://travis-ci.com/casbin/casbin)
[![Coverage Status](https://coveralls.io/repos/github/qurami/casbin-cloud-storage-adapter/badge.svg)](https://coveralls.io/github/qurami/casbin-cloud-storage-adapter)
[![Godoc](https://godoc.org/github.com/qurami/casbin-cloud-storage-adapter?status.svg)](https://pkg.go.dev/github.com/qurami/casbin-cloud-storage-adapter)

---

[Casbin](https://casbin.org/) adapter implementation for GCP Cloud Storage.
With this library, Casbin can load or save policies from/to Google Cloud Storage buckets.

## Installation

```
go get github.com/qurami/casbin-cloud-storage-adapter
```

## Example Usage

```go
package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"github.com/casbin/casbin/v2"
	cloudstorageadapter "github.com/qurami/casbin-cloud-storage-adapter"
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
```

The same file with the corresponding RBAC model is available in the [examples](examples) folder.

## Missing Features

This version is missing the _autosave_ features, so please remember to manually execute the `enforcer.SavePolicy` method when using this adapter.

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.