// (c) Copyright IBM Corp. 2024

//go:build integration
// +build integration

package instacosmos_test

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
)

func Test_pk_NewPartitionKeyString(t *testing.T) {
	type fields struct {
		value string
	}
	type args struct {
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   azcosmos.PartitionKey
	}{
		{
			name: "success",
			fields: fields{
				value: "",
			},
			args: args{
				value: "sample-partition-key",
			},
			want: azcosmos.NewPartitionKeyString("sample-partition-key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icc := instacosmos.NewPartitionKey(tt.fields.value)
			if got := icc.NewPartitionKeyString(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pk.NewPartitionKeyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pk_NewPartitionKeyBool(t *testing.T) {
	type fields struct {
		value string
	}
	type args struct {
		value bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   azcosmos.PartitionKey
	}{
		{
			name: "success",
			fields: fields{
				value: "",
			},
			args: args{
				value: true,
			},
			want: azcosmos.NewPartitionKeyBool(true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icc := instacosmos.NewPartitionKey(tt.fields.value)
			if got := icc.NewPartitionKeyBool(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pk.NewPartitionKeyBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pk_NewPartitionKeyNumber(t *testing.T) {
	type fields struct {
		value string
	}
	type args struct {
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   azcosmos.PartitionKey
	}{
		{
			name: "success",
			fields: fields{
				value: "",
			},
			args: args{
				value: 3.14,
			},
			want: azcosmos.NewPartitionKeyNumber(3.14),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icc := instacosmos.NewPartitionKey(tt.fields.value)
			if got := icc.NewPartitionKeyNumber(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pk.NewPartitionKeyNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
