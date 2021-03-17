package cloudstorageadapter

import (
	"context"
	"errors"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/fsouza/fake-gcs-server/fakestorage"
)

const (
	mockModelContent = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	mockPolicies = `p, reader, data1, read
p, writer, data1, write
p, commenter, data2, comment
g, writer, commenter
g, alice, writer
`
)

func TestNewAdapter(t *testing.T) {
	ctx := context.Background()
	bucketName := "mockBucketName"
	objectKey := "path/to/policy.csv"

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: []fakestorage.Object{
			{
				BucketName: bucketName,
				Name:       objectKey,
				Content:    []byte(""),
			},
		},
		Host: "127.0.0.1",
		Port: 65123,
	})
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	tests := []struct {
		name            string
		bucketName      string
		expectedErr     error
		expectedAdapter *Adapter
	}{
		{
			"NewAdapter returns error when bucket doesn't exist",
			"notExistingBucket",
			errors.New("storage: bucket doesn't exist"),
			nil,
		},
		{
			"NewAdapter returns an Adapter when no error is thrown",
			bucketName,
			nil,
			&Adapter{
				client:     server.Client(),
				bucketName: bucketName,
				objectKey:  objectKey,
				context:    ctx,
				mutex:      new(sync.Mutex),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			adapter, err := NewAdapter(server.Client(), testCase.bucketName, objectKey)
			if err != nil && testCase.expectedErr != nil && err.Error() != testCase.expectedErr.Error() {
				t.Errorf("error = %v, expected error %v", err, testCase.expectedErr)
			}
			if !reflect.DeepEqual(testCase.expectedAdapter, adapter) {
				t.Errorf("adapter = %v, expected adapter %v", adapter, testCase.expectedAdapter)
			}
		})
	}
}

func TestAdapter(t *testing.T) {
	ctx := context.Background()
	bucketName := "mockBucketName"
	objectKey := "path/to/policy.csv"

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: []fakestorage.Object{
			{
				BucketName: bucketName,
				Name:       objectKey,
				Content:    []byte(mockPolicies),
			},
		},
		Host: "127.0.0.1",
		Port: 65122,
	})
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	m, err := model.NewModelFromString(mockModelContent)
	if err != nil {
		t.Fatalf("cannot create model: %v", err)
	}

	a, err := NewAdapter(server.Client(), bucketName, objectKey)
	if err != nil {
		t.Fatalf("cannot create adapter: %v", err)
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		t.Fatalf("cannot create enforcer: %v", err)
	}

	expectedSubjects := []string{"reader", "writer", "commenter"}
	subjects := e.GetAllSubjects()
	if !reflect.DeepEqual(expectedSubjects, subjects) {
		t.Errorf("subjects = %v, expected subjects = %v", subjects, expectedSubjects)
	}

	expectedRolesForAlice := []string{"writer", "commenter"}
	rolesForAlice, err := e.GetImplicitRolesForUser("alice")
	if err != nil {
		t.Fatalf("cannot get implicit roles for user: %v", err)
	}
	if !reflect.DeepEqual(expectedRolesForAlice, rolesForAlice) {
		t.Errorf("roles = %v, expected roles = %v", rolesForAlice, expectedRolesForAlice)
	}

	result, err := e.AddRoleForUser("alice", "reader")
	if err != nil || result != true {
		t.Fatalf("cannot set new role for user: %v", err)
	}

	e.SavePolicy()

	fileReader, err := server.Client().Bucket(bucketName).Object(objectKey).NewReader(ctx)
	if err != nil {
		t.Fatalf("cannot read file from bucket: %v", err)
	}

	expectedPolicies := `p, reader, data1, read
p, writer, data1, write
p, commenter, data2, comment
g, writer, commenter
g, alice, writer
g, alice, reader
`
	newContent, _ := ioutil.ReadAll(fileReader)
	if expectedPolicies != string(newContent) {
		t.Errorf("file content = %v, expected file content = %v", string(newContent), expectedPolicies)
	}
}
