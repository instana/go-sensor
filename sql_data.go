// (c) Copyright IBM Corp. 2024

package instana

type database int8

const (
	sql_generic database = iota
	postgres
	mysql
	redis
	couchbase
	cosmos
)

// database names
const (
	Postgres  string = "postgres"
	MySQL     string = "mysql"
	Redis     string = "redis"
	Couchbase string = "couchbase"
	Cosmos    string = "cosmos"
	DB2       string = "db2"
)

// db Keys
const (
	pg_db_key          string = "pg"
	mysql_db_key       string = "mysql"
	redis_db_key       string = "redis"
	couchbase_db_key   string = "couchbase"
	cosmos_db_key      string = "cosmos"
	generic_sql_db_key string = "db"
)

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

var dbMap = map[string]database{
	Postgres:  postgres,
	MySQL:     mysql,
	Redis:     redis,
	Couchbase: couchbase,
	Cosmos:    cosmos,
	DB2:       sql_generic,
}

func db(dbName string) database {

	db, ok := dbMap[dbName]
	if !ok {
		return sql_generic
	}

	return db
}
