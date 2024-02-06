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

// NewPartitionKey returns instance of pk which implements PartitionKey interface
func NewPartitionKey(val string) PartitionKey {
	return &pk{
		value: val,
	}
}

// PartitionKey is the interface that wraps partition key generator functions for different types
type PartitionKey interface {
	NewPartitionKeyString(value string) azcosmos.PartitionKey
	NewPartitionKeyBool(value bool) azcosmos.PartitionKey
	NewPartitionKeyNumber(value float64) azcosmos.PartitionKey
	getPartitionKey() string
}

// NewPartitionKeyString creates a partition key with a string value.
func (p *pk) NewPartitionKeyString(value string) azcosmos.PartitionKey {
	p.value = value
	return azcosmos.NewPartitionKeyString(value)
}

// NewPartitionKeyBool creates a partition key with a boolean value.
func (p *pk) NewPartitionKeyBool(value bool) azcosmos.PartitionKey {
	p.value = strconv.FormatBool(value)
	return azcosmos.NewPartitionKeyBool(value)
}

// NewPartitionKeyNumber creates a partition key with a numeric value.
func (p *pk) NewPartitionKeyNumber(value float64) azcosmos.PartitionKey {
	p.value = strconv.FormatFloat(value, 'f', -1, 64)
	return azcosmos.NewPartitionKeyNumber(value)
}

func (p *pk) getPartitionKey() string {
	return p.value
}
