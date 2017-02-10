package gateway

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIP(t *testing.T) {
	i, err := GetDefault()
	assert.Nil(t, err)
	assert.Equal(t, getDefaultGateway(), i)
}

func getDefaultGateway() string {
	out, _ := exec.Command("/bin/sh", "-c", "/sbin/ip route | awk '/default/' | cut -d ' ' -f 3 | tr -d '\n'").Output()
	return string(out[:])
}

func BenchmarkGetIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDefault()
	}
}

func BenchmarkGetDefaultGateway(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getDefaultGateway()
	}
}
