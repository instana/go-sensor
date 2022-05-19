// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

// Package instasarama provides Instana tracing instrumentation for
// Kafka producers and consumers build on top of github.com/Shopify/sarama.
package instasarama

import (
	"os"
)

const KafkaHeaderEnvVarKey = "INSTANA_KAFKA_HEADER_FORMAT"

const (
	BINARY = "binary"
	STRING = "string"
	BOTH   = "both"
)

func getKafkaHeaderFormat() string {
	kafkaHeaderEnvVar, ok := os.LookupEnv(KafkaHeaderEnvVarKey)

	if !ok {
		kafkaHeaderEnvVar = "binary"
	}

	return kafkaHeaderEnvVar
}
