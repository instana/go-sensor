package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/instana/go-sensor/aws"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestECSMetadataProvider_ContainerMetadata(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "testdata/container_metadata.json")
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	container, err := c.ContainerMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, aws.ECSContainerMetadata{
		DockerID:      "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
		Name:          "nginx-curl",
		DockerName:    "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
		Image:         "nrdlngr/nginx-curl",
		ImageID:       "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
		DesiredStatus: "RUNNING",
		KnownStatus:   "RUNNING",
		Limits:        aws.ContainerLimits{CPU: 512, Memory: 512},
		CreatedAt:     time.Date(2018, time.February, 1, 20, 55, 10, 554941919, time.UTC),
		StartedAt:     time.Date(2018, time.February, 1, 20, 55, 11, 64236631, time.UTC),
		Type:          "NORMAL",
		Networks: []aws.ContainerNetwork{
			{Mode: "awsvpc", IPv4Addresses: []string{"10.0.2.106"}},
		},
		ContainerLabels: aws.ContainerLabels{
			Cluster:               "default",
			TaskARN:               "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
			TaskDefinition:        "nginx",
			TaskDefinitionVersion: "5",
		},
	}, container)
}

func TestECSMetadataProvider_ContainerMetadata_ServerError(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "aws error", http.StatusInternalServerError)
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.ContainerMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500 Internal Server Error")
}

func TestECSMetadataProvider_ContainerMetadata_MalformedResponse(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("here is your data"))
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.ContainerMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed")
}

func TestECSMetadataProvider_TaskMetadata(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "testdata/task_metadata.json")
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	task, err := c.TaskMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, aws.ECSTaskMetadata{
		TaskARN:          "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
		Family:           "nginx",
		Revision:         "5",
		AvailabilityZone: "us-east-2b",
		DesiredStatus:    "RUNNING",
		KnownStatus:      "RUNNING",
		Containers: []aws.ECSContainerMetadata{
			{
				DockerID:      "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c",
				Name:          "~internal~ecs~pause",
				DockerName:    "ecs-nginx-5-internalecspause-acc699c0cbf2d6d11700",
				Image:         "amazon/amazon-ecs-pause:0.1.0",
				ImageID:       "",
				DesiredStatus: "RESOURCES_PROVISIONED",
				KnownStatus:   "RESOURCES_PROVISIONED",
				Limits:        aws.ContainerLimits{CPU: 0, Memory: 0},
				CreatedAt:     time.Date(2018, time.February, 1, 20, 55, 8, 366329616, time.UTC),
				StartedAt:     time.Date(2018, time.February, 1, 20, 55, 9, 58354915, time.UTC),
				Type:          "CNI_PAUSE",
				Networks: []aws.ContainerNetwork{
					{Mode: "awsvpc", IPv4Addresses: []string{"10.0.2.106"}},
				},
				ContainerLabels: aws.ContainerLabels{
					Cluster:               "default",
					TaskARN:               "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
					TaskDefinition:        "nginx",
					TaskDefinitionVersion: "5",
				},
			},
			{
				DockerID:      "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
				Name:          "nginx-curl",
				DockerName:    "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
				Image:         "nrdlngr/nginx-curl",
				ImageID:       "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
				DesiredStatus: "RUNNING",
				KnownStatus:   "RUNNING",
				Limits:        aws.ContainerLimits{CPU: 512, Memory: 512},
				CreatedAt:     time.Date(2018, time.February, 1, 20, 55, 10, 554941919, time.UTC),
				StartedAt:     time.Date(2018, time.February, 1, 20, 55, 11, 64236631, time.UTC),
				Type:          "NORMAL",
				Networks: []aws.ContainerNetwork{
					{Mode: "awsvpc", IPv4Addresses: []string{"10.0.2.106"}},
				},
				ContainerLabels: aws.ContainerLabels{
					Cluster:               "default",
					TaskARN:               "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
					TaskDefinition:        "nginx",
					TaskDefinitionVersion: "5",
				},
			},
		},
		PullStartedAt: time.Date(2018, time.February, 1, 20, 55, 9, 372495529, time.UTC),
		PullStoppedAt: time.Date(2018, time.February, 1, 20, 55, 10, 552018345, time.UTC),
	}, task)
}

func TestECSMetadataProvider_TaskMetadata_ServerError(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "aws error", http.StatusInternalServerError)
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.TaskMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500 Internal Server Error")
}

func TestECSMetadataProvider_TaskMetadata_MalformedResponse(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("here is your data"))
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.TaskMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed")
}

func setupTS() (string, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	return srv.URL, mux, srv.Close
}
