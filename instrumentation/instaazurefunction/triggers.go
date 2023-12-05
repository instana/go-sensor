// (c) Copyright IBM Corp. 2023

package instaazurefunction

import "encoding/json"

const (
	unknownType         string = "UnknownType"
	httpTrigger         string = "HTTP"
	queueStorageTrigger string = "Queue"
)

type Metadata struct {
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
}

// Extracts the trigger type and method name from the payload
func extractSpanData(payload []byte) (Metadata, error) {
	var v struct {
		Meta Metadata `json:"MetaData"`
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return v.Meta, err
	}

	return v.Meta, nil
}

func (v Metadata) triggerName() string {
	var triggerType string

	switch {
	case v.InsertionTime != "":
		triggerType = queueStorageTrigger
	case v.Headers.UserAgent != "":
		triggerType = httpTrigger
	default:
		triggerType = unknownType
	}

	return triggerType
}

func (v Metadata) functionName() string {
	return v.Sys.MethodName
}
