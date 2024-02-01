// (c) Copyright IBM Corp. 2024

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type pk struct {
	value string
}

func NewPK(val string) *pk {
	return &pk{
		value: val,
	}
}

// PartitionKey is the interface that wraps partition key generator functions for different types
type PartitionKey interface {
	NewPartitionKeyString(value string) azcosmos.PartitionKey
	NewPartitionKeyBool(value bool) azcosmos.PartitionKey
	NewPartitionKeyNumber(value float64) azcosmos.PartitionKey
}

// NewPartitionKeyString creates a partition key with a string value.
func (icc *pk) NewPartitionKeyString(value string) azcosmos.PartitionKey {
	icc.value = value
	return azcosmos.NewPartitionKeyString(value)
}

// NewPartitionKeyBool creates a partition key with a boolean value.
func (icc *pk) NewPartitionKeyBool(value bool) azcosmos.PartitionKey {
	icc.value = strconv.FormatBool(value)
	return azcosmos.NewPartitionKeyBool(value)
}

// NewPartitionKeyNumber creates a partition key with a numeric value.
func (icc *pk) NewPartitionKeyNumber(value float64) azcosmos.PartitionKey {
	icc.value = strconv.FormatFloat(value, 'f', -1, 64)
	return azcosmos.NewPartitionKeyNumber(value)
}
