// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

// Package instasarama provides Instana tracing instrumentation for
// Kafka producers and consumers build on top of github.com/Shopify/sarama.
package instasarama

import (
	"fmt"
	"os"
	"strings"
)

const KafkaHeaderEnvVarKey = "INSTANA_KAFKA_HEADER_FORMAT"

type KafkaHeaderType int

const (
	NOT_SET = iota
	BINARY
	STRING
	BOTH
)

var kafkaHeaderTypes = map[string]KafkaHeaderType{
	"binary": BINARY,
	"string": STRING,
	"both":   BOTH,
}

func (kft KafkaHeaderType) String() string {
	switch kft {
	case BINARY:
		return "binary"
	case STRING:
		return "string"
	case BOTH:
		return "both"
	}
	return fmt.Sprintf("Unknown Kafka header type: %d", kft)
}

func GetKafkaHeaderFormat() KafkaHeaderType {
	var kafkaHeaderFormat KafkaHeaderType
	kafkaHeaderEnvVar, ok := os.LookupEnv(KafkaHeaderEnvVarKey)

	if !ok {
		kafkaHeaderEnvVar = "binary"
	}

	kafkaHeaderFormat = kafkaHeaderTypes[strings.ToLower(kafkaHeaderEnvVar)]

	if kafkaHeaderFormat == NOT_SET {
		kafkaHeaderFormat = BINARY
	}

	return kafkaHeaderFormat
}
