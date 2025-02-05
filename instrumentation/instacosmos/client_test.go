// (c) Copyright IBM Corp. 2024

//go:build integration
// +build integration

package instacosmos_test

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
	"github.com/stretchr/testify/assert"
)

func TestNewClientFromConnectionString(t *testing.T) {

	connStr := "AccountEndpoint=" + endpoint + ";AccountKey=" + key + ";"

	rec = getInstaRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    rec,
	})
	defer instana.ShutdownCollector()

	type args struct {
		collector        instana.TracerLogger
		connectionString string
		o                *azcosmos.ClientOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				collector:        c,
				connectionString: connStr,
				o:                &azcosmos.ClientOptions{},
			},
			wantErr: false,
		},
		{
			name: "	error",
			args: args{
				collector:        c,
				connectionString: "",
				o:                &azcosmos.ClientOptions{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := instacosmos.NewClientFromConnectionString(tt.args.collector, tt.args.connectionString, tt.args.o)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, endpoint == client.Endpoint())
			}
		})
	}
}

func TestNewClient(t *testing.T) {

	rec = getInstaRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    rec,
	})
	defer instana.ShutdownCollector()

	cred, err := azidentity.NewClientSecretCredential("tenantId", "clientId", "clientSecret",
		&azidentity.ClientSecretCredentialOptions{})
	assert.NoError(t, err)

	type args struct {
		collector instana.TracerLogger
		endpoint  string
		cred      azcore.TokenCredential
		o         *azcosmos.ClientOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				collector: c,
				endpoint:  endpoint,
				cred:      cred,
				o:         &azcosmos.ClientOptions{},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				collector: c,
				endpoint:  "http://{example.com",
				cred:      cred,
				o:         &azcosmos.ClientOptions{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := instacosmos.NewClient(tt.args.collector, tt.args.endpoint, tt.args.cred, tt.args.o)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, endpoint == client.Endpoint())
			}
		})
	}
}
