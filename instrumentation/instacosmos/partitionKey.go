// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type PartitionKey interface {
	NewPartitionKeyString(value string) azcosmos.PartitionKey
	NewPartitionKeyBool(value bool) azcosmos.PartitionKey
	NewPartitionKeyNumber(value float64) azcosmos.PartitionKey
}

// NewPartitionKeyString creates a partition key with a string value.
func (icc *instaContainerClient) NewPartitionKeyString(value string) azcosmos.PartitionKey {
	icc.partitionKey = value
	return azcosmos.NewPartitionKeyString(value)
}

// NewPartitionKeyBool creates a partition key with a boolean value.
func (icc *instaContainerClient) NewPartitionKeyBool(value bool) azcosmos.PartitionKey {
	icc.partitionKey = strconv.FormatBool(value)
	return azcosmos.NewPartitionKeyBool(value)
}

// NewPartitionKeyNumber creates a partition key with a numeric value.
func (icc *instaContainerClient) NewPartitionKeyNumber(value float64) azcosmos.PartitionKey {
	icc.partitionKey = strconv.FormatFloat(value, 'f', -1, 64)
	return azcosmos.NewPartitionKeyNumber(value)
}
