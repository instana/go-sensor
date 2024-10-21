// (c) Copyright IBM Corp. 2024

package instana

var redisCommands = map[string]struct{}{
	"SET":           {},
	"GET":           {},
	"DEL":           {},
	"INCR":          {},
	"DECR":          {},
	"APPEND":        {},
	"GETRANGE":      {},
	"SETRANGE":      {},
	"STRLEN":        {},
	"HSET":          {},
	"HGET":          {},
	"HMSET":         {},
	"HMGET":         {},
	"HDEL":          {},
	"HGETALL":       {},
	"HKEYS":         {},
	"HVALS":         {},
	"HLEN":          {},
	"HINCRBY":       {},
	"LPUSH":         {},
	"RPUSH":         {},
	"LPOP":          {},
	"RPOP":          {},
	"LLEN":          {},
	"LRANGE":        {},
	"LREM":          {},
	"LINDEX":        {},
	"LSET":          {},
	"SADD":          {},
	"SREM":          {},
	"SMEMBERS":      {},
	"SISMEMBER":     {},
	"SCARD":         {},
	"SINTER":        {},
	"SUNION":        {},
	"SDIFF":         {},
	"SRANDMEMBER":   {},
	"SPOP":          {},
	"ZADD":          {},
	"ZREM":          {},
	"ZRANGE":        {},
	"ZREVRANGE":     {},
	"ZRANK":         {},
	"ZREVRANK":      {},
	"ZRANGEBYSCORE": {},
	"ZCARD":         {},
	"ZSCORE":        {},
	"PFADD":         {},
	"PFCOUNT":       {},
	"PFMERGE":       {},
	"SUBSCRIBE":     {},
	"UNSUBSCRIBE":   {},
	"PUBLISH":       {},
	"MULTI":         {},
	"EXEC":          {},
	"DISCARD":       {},
	"WATCH":         {},
	"UNWATCH":       {},
	"KEYS":          {},
	"EXISTS":        {},
	"EXPIRE":        {},
	"TTL":           {},
	"PERSIST":       {},
	"RENAME":        {},
	"RENAMENX":      {},
	"TYPE":          {},
	"SCAN":          {},
	"PING":          {},
	"INFO":          {},
	"CLIENT LIST":   {},
	"CONFIG GET":    {},
	"CONFIG SET":    {},
	"FLUSHDB":       {},
	"FLUSHALL":      {},
	"DBSIZE":        {},
	"SAVE":          {},
	"BGSAVE":        {},
	"BGREWRITEAOF":  {},
	"SHUTDOWN":      {},
}
