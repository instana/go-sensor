// (c) Copyright IBM Corp. 2025

package main

import (
	"fmt"
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
)

// defining a slice named matchers to hold multiple instana.Matcher instances
type matchers []instana.Matcher

// implementing the instana.Matcher interface
func (ma matchers) Match(s string) bool {
	for _, m := range ma {
		if m.Match(s) {
			return true
		}
	}
	return false
}

// constructs a MultiMatcher consisting of:
// - an EqualsIgnoreCaseMatcher for the string "key"
// - a ContainsIgnoreCaseMatcher for the string "pass"
func CreateMultiSecretsMatcher() instana.Matcher {
	m1, _ := instana.NamedMatcher(instana.EqualsIgnoreCaseMatcher, []string{"key"})
	m2, _ := instana.NamedMatcher(instana.ContainsIgnoreCaseMatcher, []string{"pass"})
	return matchers{m1, m2}
}

func main() {
	col := instana.InitCollector(&instana.Options{
		Service:           "Nithin HTTP Secret Matcher Example",
		EnableAutoProfile: true,
		Tracer: instana.TracerOptions{
			Secrets: CreateMultiSecretsMatcher(),
		},
	})

	http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
		var msg string
		status := http.StatusOK
		client := &http.Client{
			Transport: instana.RoundTripper(col, nil),
		}

		ctx := r.Context()

		// this request contains two query parameters: keys and password
		// "keys" will not be masked, as the matcher expects the exact string "key"
		// "password" will be masked as it matches the "pass" substring in the matcher
		clientReq, err := http.NewRequest(http.MethodGet, "https://www.example.com/instana_matcher?keys=true&password=true", nil)
		if err != nil {
			status = http.StatusInternalServerError
			err := fmt.Errorf("failed to create request: %s", err)
			msg = err.Error()
		}

		_, err = client.Do(clientReq.WithContext(ctx))
		if err != nil {
			status = http.StatusInternalServerError
			err := fmt.Errorf("failed to GET https://www.example.com/instana_matcher?keys=true&password=true: %s", err)
			msg = err.Error()
		}

		msg = "Request to https://www.example.com/instana_matcher?keys=true&password=true was successful"

		w.WriteHeader(status)
		w.Write([]byte(msg))
	}))

	log.Fatal(http.ListenAndServe(":7070", nil))
}
