// (c) Copyright IBM Corp. 2023

//go:build go1.16
// +build go1.16

// Package instagorm provides instrumentation for the gorm library.
package instagorm

import (
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"gorm.io/gorm"
)

type wrappedDB struct {
	connDetails instana.DbConnDetails
	sensor      instana.TracerLogger
	*gorm.DB
}

// Instrument adds instrumentation for the specified gorm database instance.
func Instrument(db *gorm.DB, s instana.TracerLogger, dsn string) {

	wdB := wrappedDB{
		connDetails: instana.ParseDBConnDetails(dsn),
		sensor:      s,
		DB:          db,
	}

	wdB.registerCreateCallbacks()

	wdB.registerQueryCallbacks()

	wdB.registerRowCallbacks()

	wdB.registerRawCallbacks()

	wdB.registerDeleteCallbacks()

	wdB.registerUpdateCallbacks()

}

func (wdB *wrappedDB) registerCreateCallbacks() {
	wdB.logError(wdB.Callback().Create().Before("gorm:create").Register("instagorm:before_create",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Create().After("gorm:create").Register("instagorm:after_create",
		postOpCb()))
}

func (wdB *wrappedDB) registerUpdateCallbacks() {
	wdB.logError(wdB.Callback().Update().Before("gorm:update").Register("instagorm:before_update",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Update().After("gorm:update").Register("instagorm:after_update",
		postOpCb()))
}

func (wdB *wrappedDB) registerDeleteCallbacks() {
	wdB.logError(wdB.Callback().Delete().Before("gorm:delete").Register("instagorm:before_delete",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Delete().After("gorm:delete").Register("instagorm:after_delete",
		postOpCb()))

}

func (wdB *wrappedDB) registerQueryCallbacks() {
	wdB.logError(wdB.Callback().Query().Before("gorm:query").Register("instagorm:before_query",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Query().After("gorm:query").Register("instagorm:after_query",
		postOpCb()))

}

func (wdB *wrappedDB) registerRowCallbacks() {
	wdB.logError(wdB.Callback().Raw().Before("gorm:row").Register("instagorm:before_row",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Raw().After("gorm:row").Register("instagorm:after_row",
		postOpCb()))
}

func (wdB *wrappedDB) registerRawCallbacks() {
	wdB.logError(wdB.Callback().Raw().Before("gorm:raw").Register("instagorm:before_raw",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Raw().After("gorm:raw").Register("instagorm:after_raw",
		postOpCb()))
}

func (wdB *wrappedDB) logError(err error) {
	if err != nil {
		wdB.sensor.Logger().Error("unable to register callback, error: ", err.Error())
	}
}

func preOpCb(wdB *wrappedDB) func(db *gorm.DB) {

	return func(db *gorm.DB) {

		ctx := db.Statement.Context

		sp, dbKey := instana.StartSQLSpan(ctx, wdB.connDetails, db.Statement.SQL.String(), wdB.sensor)

		sp.SetBaggageItem("dbKey", dbKey)

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

		dbKey := sp.BaggageItem("dbKey")

		var stmtKey string = dbKey + ".stmt"
		var errKey string = dbKey + ".error"

		if dbKey == "db" {
			stmtKey = string(ext.DBStatement)
			errKey = "error"
			sp.SetTag(string(ext.DBType), db.Dialector.Name())
		}

		sp.SetTag(stmtKey, db.Statement.SQL.String())

		if err := db.Statement.Error; err != nil {
			sp.SetTag(errKey, err.Error())
			sp.LogFields(otlog.Error(err))
		}
	}
}
