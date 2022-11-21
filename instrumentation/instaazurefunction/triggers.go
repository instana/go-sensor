// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

//go:build go1.13
// +build go1.13

package instaazurefunction

import "encoding/json"

const (
	unknownType string = "unknown type"

	//supported types
	httpTrigger         string = "httpTrigger"
	queueStorageTrigger string = "queueTrigger"
)

// Extracts the trigger type and method name from the payload
func extractSpanData(payload []byte) (string, string) {
	var v struct {
		Metadata struct {
			//queueStorage fields
			DequeueCount int    `json:"DequeueCount"`
			PopReceipt   string `json:"PopReceipt"`
			//http fields
			Headers struct {
				UserAgent string `json:"User-Agent"`
			} `json:"Headers"`
			//common info with method name
			Sys struct {
				MethodName string `json:"MethodName"`
			} `json:"sys"`
		} `json:"MetaData"`
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return unknownType, ""
	}

	method := v.Metadata.Sys.MethodName

	switch {
	case v.Metadata.PopReceipt != "":
		return queueStorageTrigger, method
	case v.Metadata.Headers.UserAgent != "":
		return httpTrigger, method
	default:
		return unknownType, method
	}
}
