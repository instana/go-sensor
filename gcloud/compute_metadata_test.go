// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package gcloud_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/instana/go-sensor/gcloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeMetadataProvider_ComputeMetadata(t *testing.T) {
	baseURL, mux, teardown := setupTS()
	defer teardown()

	data, err := ioutil.ReadFile("testdata/computeMetadata.json")
	require.NoError(t, err)

	mux.HandleFunc("/computeMetadata/v1", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "true", req.URL.Query().Get("recursive"))
		assert.Equal(t, "Google", req.Header.Get("Metadata-Flavor"))

		w.Write(data)
	})

	c := gcloud.NewComputeMetadataProvider(baseURL, nil)

	md, err := c.ComputeMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, gcloud.ComputeMetadata{
		Project: gcloud.ProjectMetadata{
			NumericProjectID: 1234567890,
			ProjectID:        "test-project",
		},
		Instance: gcloud.InstanceMetadata{
			ID:     "id1",
			Region: "projects/1234567890/regions/us-central1",
		},
	}, md)
}

func setupTS() (string, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	return srv.URL, mux, srv.Close
}
