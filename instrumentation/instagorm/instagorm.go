// (c) Copyright IBM Corp. 2023

//go:build go1.16
// +build go1.16

// Package instagorm provides instrumentation for the gorm library.
package instagorm

import (
	"encoding/json"

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
	config dbConnectionConfig
	sensor *instana.Sensor
	db     *gorm.DB
}

// Instrument adds instrumentation for the specified gorm database instance.
func Instrument(db *gorm.DB, s *instana.Sensor, dsn string) {

	wdB := wrappedDB{
		config: parseDSN(dsn),
		sensor: s,
		db:     db,
	}

	wdB.registerCreateCallbacks()

	wdB.queryCallbacks()

	wdB.rowCallbacks()

	wdB.rawCallbacks()

	wdB.deleteCallbacks()

	wdB.updateCallbacks()

}

func parseDSN(dsn string) dbConnectionConfig {
	var cfg dbConnectionConfig
	var cfgStr string

	//TODO: Replace with instana.ParseDbConnectionDetails
	if cfgStr = instana.GetDBConnectDetails(dsn); cfgStr == "" {
		return dbConnectionConfig{RawString: dsn}
	}

	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return dbConnectionConfig{RawString: dsn}
	}

	return cfg
}

func (wdB *wrappedDB) registerCreateCallbacks() {

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Create().Before("gorm:before_create").Register("instagorm:before_create",
		wdB.startCb()))

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Create().After("gorm:after_create").Register("instagorm:after_create",
		wdB.stopCb()))

}

func (wdB *wrappedDB) updateCallbacks() {

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Update().Before("gorm:before_update").Register("instagorm:before_update",
		wdB.startCb()))

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Update().After("gorm:after_update").Register("instagorm:after_update",
		wdB.stopCb()))
}

func (wdB *wrappedDB) deleteCallbacks() {

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Delete().After("gorm:before_delete").Register("instagorm:before_delete",
		wdB.startCb()))

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Delete().After("gorm:after_delete").Register("instagorm:after_delete",
		wdB.stopCb()))

}

func (wdB *wrappedDB) queryCallbacks() {

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Query().Before("gorm:query").Register("instagorm:before_query",
		wdB.startCb()))

	wdB.checkForSuccessfulRegn(wdB.db.Callback().Query().After("gorm:after_query").Register("instagorm:after_query",
		wdB.stopCb()))

}

func (wdB *wrappedDB) rowCallbacks() {
	wdB.checkForSuccessfulRegn(wdB.db.Callback().Raw().Before("gorm:row").Register("instagorm:before_row",
		wdB.startCb()))
	wdB.checkForSuccessfulRegn(wdB.db.Callback().Raw().After("gorm:row").Register("instagorm:after_row",
		wdB.stopCb()))
}

func (wdB *wrappedDB) rawCallbacks() {
	wdB.checkForSuccessfulRegn(wdB.db.Callback().Raw().Before("gorm:raw").Register("instagorm:before_raw",
		wdB.startCb()))
	wdB.checkForSuccessfulRegn(wdB.db.Callback().Raw().After("gorm:raw").Register("instagorm:after_raw",
		wdB.stopCb()))
}

func (wdB *wrappedDB) checkForSuccessfulRegn(err error) {
	if err != nil {
		wdB.sensor.Logger().Error("unable to register callback for gorm create")
	}
}

func (wdB *wrappedDB) startCb() func(db *gorm.DB) {
	return func(db *gorm.DB) {

		wdB.startSpan()
	}
}

func (wdB *wrappedDB) stopCb() func(db *gorm.DB) {
	return func(db *gorm.DB) {

		wdB.stopSpan(db)
	}
}

func (wdB *wrappedDB) startSpan() {
	var sp ot.Span

	ctx := wdB.db.Statement.Context

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

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := instana.SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	sp = wdB.sensor.Tracer().StartSpan("sdk.database", opts...)
	ctx = instana.ContextWithSpan(ctx, sp)
	wdB.db.Statement.Context = ctx
}

func (wdB *wrappedDB) stopSpan(db *gorm.DB) {
	sp, ok := instana.SpanFromContext(wdB.db.Statement.Context)
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
