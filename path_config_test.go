package balenakeys

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

const (
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NzE5MjAsInVzZXJuYW1lIjoiYWRtaW41MyIsImp3dF9zZWNyZXQiOiJLNlJZRzc0WkdQS1IzNkhST0ZSSFVNVElJQkpFWVVTTSIsInR3b0ZhY3RvclJlcXVpcmVkIjpmYWxzZSwiYXV0aFRpbWUiOjE2OTM5NDU4Njk5MjcsImlzX3ZlcmlmaWVkIjp0cnVlLCJtdXN0X2JlX3ZlcmlmaWVkIjpmYWxzZSwiYWN0b3IiOnsiX19pZCI6Mzg0MzA4Nn0sInBlcm1pc3Npb25zIjpbXSwiZW1haWwiOiJhZG1pbkBzbWFydHZpc2lvbndvcmtzLmNvbSIsImNyZWF0ZWRfYXQiOiIyMDE5LTA1LTAyVDEzOjM0OjM4LjkxN1oiLCJoYXNfZGlzYWJsZWRfbmV3c2xldHRlciI6dHJ1ZSwiZmlyc3RfbmFtZSI6IkN1cnRpcyIsImxhc3RfbmFtZSI6IktvZWxsaW5nIiwiYWNjb3VudF90eXBlIjoicHJvZmVzc2lvbmFsIiwic29jaWFsX3NlcnZpY2VfYWNjb3VudCI6W10sImNvbXBhbnkiOiJTbWFydCBWaXNpb24gV29ya3MiLCJoYXNQYXNzd29yZFNldCI6dHJ1ZSwicHVibGljX2tleSI6dHJ1ZSwiZmVhdHVyZXMiOltdLCJpbnRlcmNvbVVzZXJOYW1lIjoiYWRtaW41MyIsImludGVyY29tVXNlckhhc2giOiIwYmZmZGIwMTA1Mzc2NDEzOTkyODkyNTAxMTY1ZmI3NThhMDFjMTJkMWMyYzVjODgyZWEwNWFmOWJlZGM4NzQ5IiwiaWF0IjoxNjk1MTQwMTQwLCJleHAiOjE2OTU3NDQ5NDB9.tVjxtSQ_dzTmBlHFRrfvDKFrGEgzUXjzSZC_0l7qFOs"
	url   = "https://api.balena-cloud.com/"
)

// TestConfig mocks the creation, read, update, and delete
// of the backend configuration for HashiCups.
func TestConfig(t *testing.T) {
	b, reqStorage := getTestBackend(t)

	t.Run("Test Configuration", func(t *testing.T) {
		err := testConfigCreate(t, b, reqStorage, map[string]interface{}{
			"token": token,
			"url":   url,
		})

		assert.NoError(t, err)

		err = testConfigRead(t, b, reqStorage, map[string]interface{}{
			"url": url,
		})

		assert.NoError(t, err)

		err = testConfigUpdate(t, b, reqStorage, map[string]interface{}{
			"token": token,
			"url":   url,
		})

		assert.NoError(t, err)

		err = testConfigRead(t, b, reqStorage, map[string]interface{}{
			"url": url,
		})

		assert.NoError(t, err)

		err = testConfigDelete(t, b, reqStorage)

		assert.NoError(t, err)
	})
}

func testConfigDelete(t *testing.T, b logical.Backend, s logical.Storage) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "config",
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigCreate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Data:      d,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      d,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp == nil && expected == nil {
		return nil
	}

	if resp.IsError() {
		return resp.Error()
	}

	if len(expected) != len(resp.Data) {
		return fmt.Errorf("read data mismatch (expected %d values, got %d)", len(expected), len(resp.Data))
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]

		if !ok {
			return fmt.Errorf(`expected data["%s"] = %v but was not included in read output"`, k, expectedV)
		} else if expectedV != actualV {
			return fmt.Errorf(`expected data["%s"] = %v, instead got %v"`, k, expectedV, actualV)
		}
	}

	return nil
}
