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

type dbConnectionConfig struct {
	RawString  string
	Host, Port string
	Schema     string
	User       string
}

type dbManager struct {
	config dbConnectionConfig
	sensor *instana.Sensor
	db     *gorm.DB
}

// Instrument adds instrumentation for the specified gorm database instance.
func Instrument(db *gorm.DB, s *instana.Sensor, dsn string) {

	dbMgr := dbManager{
		config: parseDSN(dsn),
		sensor: s,
		db:     db,
	}

	dbMgr.registerCreateCallbacks()

	dbMgr.queryCallbacks()

	dbMgr.rowCallbacks()

	dbMgr.rawCallbacks()

	dbMgr.deleteCallbacks()

	dbMgr.updateCallbacks()

}

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

func (dbMgr *dbManager) registerCreateCallbacks() {

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Create().Before("gorm:before_create").Register("instagorm:before_create",
		dbMgr.startCb()))

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Create().After("gorm:after_create").Register("instagorm:after_create",
		dbMgr.stopCb()))

}

func (dbMgr *dbManager) updateCallbacks() {

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Update().Before("gorm:before_update").Register("instagorm:before_update",
		dbMgr.startCb()))

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Update().After("gorm:after_update").Register("instagorm:after_update",
		dbMgr.stopCb()))
}

func (dbMgr *dbManager) deleteCallbacks() {

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Delete().After("gorm:before_delete").Register("instagorm:before_delete",
		dbMgr.startCb()))

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Delete().After("gorm:after_delete").Register("instagorm:after_delete",
		dbMgr.stopCb()))

}

func (dbMgr *dbManager) queryCallbacks() {

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Query().Before("gorm:query").Register("instagorm:before_query",
		dbMgr.startCb()))

	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Query().After("gorm:after_query").Register("instagorm:after_query",
		dbMgr.stopCb()))

}

func (dbMgr *dbManager) rowCallbacks() {
	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Raw().Before("gorm:row").Register("instagorm:before_row",
		dbMgr.startCb()))
	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Raw().After("gorm:row").Register("instagorm:after_row",
		dbMgr.stopCb()))
}

func (dbMgr *dbManager) rawCallbacks() {
	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Raw().Before("gorm:raw").Register("instagorm:before_raw",
		dbMgr.startCb()))
	dbMgr.checkForSuccessfulRegn(dbMgr.db.Callback().Raw().After("gorm:raw").Register("instagorm:after_raw",
		dbMgr.stopCb()))
}

func (dbMgr *dbManager) checkForSuccessfulRegn(err error) {
	if err != nil {
		dbMgr.sensor.Logger().Error("unable to register callback for gorm create")
	}
}

func (dbMgr *dbManager) startCb() func(db *gorm.DB) {
	return func(db *gorm.DB) {

		dbMgr.startSpan()
	}
}

func (dbMgr *dbManager) stopCb() func(db *gorm.DB) {
	return func(db *gorm.DB) {

		dbMgr.stopSpan(db)
	}
}

func (dbMgr *dbManager) startSpan() {
	var sp ot.Span

	ctx := dbMgr.db.Statement.Context

	tags := ot.Tags{
		string(ext.DBType):      "sql",
		string(ext.DBStatement): dbMgr.db.Statement.SQL.String(),
		string(ext.PeerAddress): dbMgr.config.RawString,
	}

	if dbMgr.config.Schema != "" {
		tags[string(ext.DBInstance)] = dbMgr.config.Schema
	} else {
		tags[string(ext.DBInstance)] = dbMgr.config.RawString
	}

	if dbMgr.config.Host != "" {
		tags[string(ext.PeerHostname)] = dbMgr.config.Host
	}

	if dbMgr.config.Port != "" {
		tags[string(ext.PeerPort)] = dbMgr.config.Port
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := instana.SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	sp = dbMgr.sensor.Tracer().StartSpan("sdk.database", opts...)
	ctx = instana.ContextWithSpan(ctx, sp)
	dbMgr.db.Statement.Context = ctx
}

func (dbMgr *dbManager) stopSpan(db *gorm.DB) {
	sp, ok := instana.SpanFromContext(dbMgr.db.Statement.Context)
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
