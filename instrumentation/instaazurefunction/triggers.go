// (c) Copyright IBM Corp. 2023

package instaazurefunction

import "encoding/json"

const (
	unknownType         string = "UnknownType"
	httpTrigger         string = "HTTP"
	queueStorageTrigger string = "Queue"
)

type spanData struct {
	FunctionName string
	TriggerType  string
}

// Extracts the trigger type and method name from the payload
func extractSpanData(payload []byte) (spanData, error) {
	var s spanData
	var v struct {
		Metadata struct {
			//queueStorage fields
			DequeueCount  int    `json:"DequeueCount,string"`
			PopReceipt    string `json:"PopReceipt"`
			InsertionTime string `json:"InsertionTime"`
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

	s.FunctionName = v.Metadata.Sys.MethodName

	switch {
	case v.Metadata.InsertionTime != "":
		s.TriggerType = queueStorageTrigger
	case v.Metadata.Headers.UserAgent != "":
		s.TriggerType = httpTrigger
	default:
		s.TriggerType = unknownType
	}

	return s, nil
}
