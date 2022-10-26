// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2018

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

const eumExpectedResult string = `<script>
(function(c,e,f,k,g,h,b,a,d){c[g]||(c[g]=h,b=c[h]=function(){b.q.push(arguments)},b.q=[],b.l=1*new Date,a=e.createElement(f),a.async=1,a.src=k,a.setAttribute("crossorigin","anonymous"),d=e.getElementsByTagName(f)[0],d.parentNode.insertBefore(a,d))})(window,document,"script","//eum.instana.io/eum.min.js","InstanaEumObject","ineum");ineum('reportingUrl','https://eum-saas.instana.io');ineum('key','myApiKey');ineum('traceId','myTraceId');ineum('meta','key1','value1');ineum('meta','key2','value2');
</script>
`

func TestEum(t *testing.T) {
	assert.Equal(t, eumExpectedResult, instana.EumSnippet("myApiKey", "myTraceId", map[string]string{
		"key1": "value1",
		"key2": "value2",
	}))
}
