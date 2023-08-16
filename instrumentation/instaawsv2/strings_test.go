// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: This file contains the testcases for testing the private methods of the package.

var (
	placeHolder      string = "test-string"
	placeHolderEmpty string = ""
)

func TestStringDeRef(t *testing.T) {
	testcases := map[string]struct {
		stringAddr  *string
		expectedVal string
	}{
		"Valid string addr": {
			stringAddr:  &placeHolder,
			expectedVal: "test-string",
		},
		"Nil addr": {
			stringAddr:  nil,
			expectedVal: "",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			actual := stringDeRef(testcase.stringAddr)
			assert.Equal(t, testcase.expectedVal, actual)
		})
	}
}

func TestStringRef(t *testing.T) {
	testcases := map[string]struct {
		input        *string
		expectedAddr *string
	}{
		"Valid string addr": {
			input:        &placeHolder,
			expectedAddr: &placeHolder,
		},
		"Nil addr": {
			input:        &placeHolderEmpty,
			expectedAddr: &placeHolderEmpty,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			addr := stringRef(*testcase.input)
			assert.Equal(t, testcase.expectedAddr, addr)
		})
	}
}
