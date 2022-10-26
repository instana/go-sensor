// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestECSMetadataProvider_TaskStats(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task/stats", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "testdata/task_stats.json")
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	data, err := c.TaskStats(context.Background())
	require.NoError(t, err)

	require.Contains(t, data, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946")

	stats := data["43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946"]
	stats.ReadAt = stats.ReadAt.UTC().Truncate(time.Second)

	assert.Equal(t, docker.ContainerStats{
		ReadAt: time.Date(2020, time.September, 9, 9, 54, 21, 0, time.UTC),
		Networks: map[string]docker.ContainerNetworkStats{
			"eth1": {
				TxDropped: 1,
				TxErrors:  2,
				TxPackets: 444,
				TxBytes:   106367,
				RxDropped: 3,
				RxErrors:  4,
				RxPackets: 3105,
				RxBytes:   4172695,
			},
		},
		Memory: docker.ContainerMemoryStats{
			Stats: docker.MemoryStats{
				ActiveAnon:   4681728,
				ActiveFile:   12288,
				InactiveAnon: 100,
				InactiveFile: 602112,
				TotalRss:     5283840,
				TotalCache:   12290,
			},
			MaxUsage: 6549504,
			Usage:    6148096,
			Limit:    536870912,
		},
		CPU: docker.ContainerCPUStats{
			System:     360200000000,
			OnlineCPUs: 2,
			Usage: docker.CPUUsageStats{
				Total:  281318382,
				Kernel: 20000000,
				User:   180000000,
			},
			Throttling: docker.CPUThrottlingStats{
				Periods: 1,
				Time:    3,
			},
		},
		BlockIO: docker.ContainerBlockIOStats{
			ServiceBytes: []docker.BlockIOOpStats{
				{Operation: docker.BlockIOReadOp, Value: 1},
				{Operation: docker.BlockIOWriteOp, Value: 2},
				{Operation: docker.BlockIOReadOp, Value: 3},
			},
		},
	}, stats)
}

func TestECSMetadataProvider_TaskStats_ServerError(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task/stats", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "aws error", http.StatusInternalServerError)
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.TaskStats(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500 Internal Server Error")
}

func TestECSMetadataProvider_TaskStats_MalformedResponse(t *testing.T) {
	endpoint, mux, teardown := setupTS()
	defer teardown()

	mux.HandleFunc("/task/stats", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("here is your data"))
	})

	c := aws.NewECSMetadataProvider(endpoint, nil)

	_, err := c.TaskStats(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed")
}

func setupTS() (string, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	return srv.URL, mux, srv.Close
}
