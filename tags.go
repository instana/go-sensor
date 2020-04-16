package instana

import ot "github.com/opentracing/opentracing-go"

const batchSizeTag = "batch_size"

// BatchSize returns an opentracing.Tag to mark the span as a batched span representing
// similar span categories. An example of such span would be batch writes to a queue,
// a database, etc. If the batch size less than 2, then this option has no effect
func BatchSize(n int) ot.Tag {
	return ot.Tag{Key: batchSizeTag, Value: n}
}
