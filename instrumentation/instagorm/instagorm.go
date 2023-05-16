// (c) Copyright IBM Corp. 2023

//go:build go1.16
// +build go1.16

// Package instagorm provides instrumentation for the gorm library.
package instagorm

import (
	"encoding/json"
	"sync"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"gorm.io/gorm"
)

// TODO: Remove the redundant struct
type dbConnectionConfig struct {
	RawString  string
	Host, Port string
	Schema     string
	User       string
}

type wrappedDB struct {
	// TODO: Change config to instana.DbConnectionDetails
	config dbConnectionConfig
	sensor *instana.Sensor
	db     *gorm.DB
	mu     sync.Mutex
}

// Instrument adds instrumentation for the specified gorm database instance.
func Instrument(db *gorm.DB, s *instana.Sensor, dsn string) {

	wdB := wrappedDB{
		// TODO: Replace with instana.ParseDbConnectionDetails.
		config: parseDSN(dsn),
		sensor: s,
		db:     db,
	}

	wdB.registerCreateCallbacks()

	wdB.registerQueryCallbacks()

	wdB.registerRowCallbacks()

	wdB.registerRawCallbacks()

	wdB.registerDeleteCallbacks()

	wdB.registerUpdateCallbacks()

}

// TODO: Remove the redundant function.
func parseDSN(dsn string) dbConnectionConfig {
	var cfg dbConnectionConfig
	var cfgStr string

	if cfgStr = instana.GetDBConnectDetails(dsn); cfgStr == "" {
		return dbConnectionConfig{RawString: dsn}
	}

	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return dbConnectionConfig{RawString: dsn}
	}

	return cfg
}

func (wdB *wrappedDB) registerCreateCallbacks() {
	wdB.logError(wdB.db.Callback().Create().Before("gorm:before_create").Register("instagorm:before_create",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Create().After("gorm:after_create").Register("instagorm:after_create",
		postOpCb()))
}

func (wdB *wrappedDB) registerUpdateCallbacks() {
	wdB.logError(wdB.db.Callback().Update().Before("gorm:before_update").Register("instagorm:before_update",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Update().After("gorm:after_update").Register("instagorm:after_update",
		postOpCb()))
}

func (wdB *wrappedDB) registerDeleteCallbacks() {
	wdB.logError(wdB.db.Callback().Delete().After("gorm:before_delete").Register("instagorm:before_delete",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Delete().After("gorm:after_delete").Register("instagorm:after_delete",
		postOpCb()))

}

func (wdB *wrappedDB) registerQueryCallbacks() {
	wdB.logError(wdB.db.Callback().Query().Before("gorm:query").Register("instagorm:before_query",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Query().After("gorm:after_query").Register("instagorm:after_query",
		postOpCb()))

}

func (wdB *wrappedDB) registerRowCallbacks() {
	wdB.logError(wdB.db.Callback().Raw().Before("gorm:row").Register("instagorm:before_row",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Raw().After("gorm:row").Register("instagorm:after_row",
		postOpCb()))
}

func (wdB *wrappedDB) registerRawCallbacks() {
	wdB.logError(wdB.db.Callback().Raw().Before("gorm:raw").Register("instagorm:before_raw",
		preOpCb(wdB)))

	wdB.logError(wdB.db.Callback().Raw().After("gorm:raw").Register("instagorm:after_raw",
		postOpCb()))
}

func (wdB *wrappedDB) logError(err error) {
	if err != nil {
		wdB.sensor.Logger().Error("unable to register callback, error: ", err.Error())
	}
}

func preOpCb(wdB *wrappedDB) func(db *gorm.DB) {

	return func(db *gorm.DB) {

		var sp ot.Span

		ctx := db.Statement.Context

		tags := wdB.generateTags()

		opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
		if parentSpan, ok := instana.SpanFromContext(ctx); ok {
			opts = append(opts, ot.ChildOf(parentSpan.Context()))
		}

		sp = wdB.sensor.Tracer().StartSpan("sdk.database", opts...)
		ctx = instana.ContextWithSpan(ctx, sp)
		db.Statement.Context = ctx
	}
}

func postOpCb() func(db *gorm.DB) {

	return func(db *gorm.DB) {
		sp, ok := instana.SpanFromContext(db.Statement.Context)
		if !ok {
			return
		}

		defer sp.Finish()

		sp.SetTag(string(ext.DBStatement), db.Statement.SQL.String())

		if err := db.Statement.Error; err != nil {
			sp.SetTag("error", err.Error())
			sp.LogFields(otlog.Error(err))
		}
	}
}

func (wdB *wrappedDB) generateTags() ot.Tags {
	wdB.mu.Lock()
	defer wdB.mu.Unlock()

	tags := ot.Tags{
		string(ext.DBType):      "sql",
		string(ext.DBStatement): wdB.db.Statement.SQL.String(),
		string(ext.PeerAddress): wdB.config.RawString,
	}

	if wdB.config.Schema != "" {
		tags[string(ext.DBInstance)] = wdB.config.Schema
	} else {
		tags[string(ext.DBInstance)] = wdB.config.RawString
	}

	if wdB.config.Host != "" {
		tags[string(ext.PeerHostname)] = wdB.config.Host
	}

	if wdB.config.Port != "" {
		tags[string(ext.PeerPort)] = wdB.config.Port
	}

	return tags
}
