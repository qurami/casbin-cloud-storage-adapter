package cloudstorageadapter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
)

// Adapter implements casbin/persist.Adapter
// storing policy configuration on a Google Cloud Storage Bucket
type Adapter struct {
	client     *storage.Client
	bucketName string
	objectKey  string
	context    context.Context
	mutex      *sync.Mutex
}

// NewAdapter creates new Adapter
//
// Parameters:
// - client
//     A cloud.google.com/go/storage.Client object
// - bucketName
//     Name of the bucket where the policy configuration file is stored on
// - objectKey
//     Key (name) of the object that contains policy configuration
func NewAdapter(client *storage.Client, bucketName string, objectKey string) (*Adapter, error) {
	ctx := context.Background()
	_, err := client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	adapter := Adapter{
		client:     client,
		bucketName: bucketName,
		objectKey:  objectKey,
		context:    ctx,
		mutex:      new(sync.Mutex),
	}
	return &adapter, nil
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model model.Model) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	fileReader, err := a.client.Bucket(a.bucketName).Object(a.objectKey).NewReader(a.context)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	buf := bufio.NewReader(fileReader)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		persist.LoadPolicyLine(line, model)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	fileWriter := a.client.Bucket(a.bucketName).Object(a.objectKey).NewWriter(a.context)
	defer fileWriter.Close()

	for ptype, assertion := range model["p"] {
		for _, rule := range assertion.Policy {
			_, err := fileWriter.Write([]byte(fmt.Sprintf("%s, %s\n", ptype, util.ArrayToString(rule))))
			if err != nil {
				return err
			}
		}
	}

	for ptype, assertion := range model["g"] {
		for _, rule := range assertion.Policy {
			_, err := fileWriter.Write([]byte(fmt.Sprintf("%s, %s\n", ptype, util.ArrayToString(rule))))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}
