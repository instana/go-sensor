// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

package instaazurefunction

import "encoding/json"

const (
	unknownType         string = "unknown type"
	httpTrigger         string = "httpTrigger"
	queueStorageTrigger string = "queueTrigger"
)

type spanData struct {
	MethodName  string
	TriggerType string
}

// Extracts the trigger type and method name from the payload
func extractSpanData(payload []byte) (spanData, error) {
	var s spanData
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
		return s, err
	}

	s.MethodName = v.Metadata.Sys.MethodName

	switch {
	case v.Metadata.PopReceipt != "":
		s.TriggerType = queueStorageTrigger
	case v.Metadata.Headers.UserAgent != "":
		s.TriggerType = httpTrigger
	default:
		s.TriggerType = unknownType
	}

	return s, nil
}
