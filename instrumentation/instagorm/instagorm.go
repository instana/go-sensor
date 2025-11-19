// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

// Package instagorm provides instrumentation for the gorm library.
package instagorm

import (
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/logger"
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

	if !isValidParams(db, s, dsn) {
		return
	}

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

func isValidParams(db *gorm.DB, s instana.TracerLogger, dsn string) bool {

	var ok bool = true

	if s == nil {
		ok = false

		var logger instana.LeveledLogger = logger.New(nil)
		logger.Error("instagorm: cannot instrument gorm.DB instance without tracer")
		return ok
	}

	if s.Logger() == nil {
		var logger instana.LeveledLogger = logger.New(nil)
		s.SetLogger(logger)
	}

	if db == nil {
		ok = false

		if s.Logger() != nil {
			s.Logger().Error("instagorm: cannot instrument nil gorm.DB instance")
		}
	}

	if dsn == "" {
		s.Logger().Warn("instagorm: received empty DSN while instrumenting gorm.DB instance")
	}

	return ok
}

func (wdB *wrappedDB) registerCreateCallbacks() {
	wdB.logError(wdB.Callback().Create().Before("gorm:create").Register("instagorm:before_create",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Create().After("gorm:create").Register("instagorm:after_create",
		postOpCb(wdB)))
}

func (wdB *wrappedDB) registerUpdateCallbacks() {
	wdB.logError(wdB.Callback().Update().Before("gorm:update").Register("instagorm:before_update",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Update().After("gorm:update").Register("instagorm:after_update",
		postOpCb(wdB)))
}

func (wdB *wrappedDB) registerDeleteCallbacks() {
	wdB.logError(wdB.Callback().Delete().Before("gorm:delete").Register("instagorm:before_delete",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Delete().After("gorm:delete").Register("instagorm:after_delete",
		postOpCb(wdB)))

}

func (wdB *wrappedDB) registerQueryCallbacks() {
	wdB.logError(wdB.Callback().Query().Before("gorm:query").Register("instagorm:before_query",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Query().After("gorm:query").Register("instagorm:after_query",
		postOpCb(wdB)))

}

func (wdB *wrappedDB) registerRowCallbacks() {
	wdB.logError(wdB.Callback().Row().Before("gorm:row").Register("instagorm:before_row",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Row().After("gorm:row").Register("instagorm:after_row",
		postOpCb(wdB)))
}

func (wdB *wrappedDB) registerRawCallbacks() {
	wdB.logError(wdB.Callback().Raw().Before("gorm:raw").Register("instagorm:before_raw",
		preOpCb(wdB)))

	wdB.logError(wdB.Callback().Raw().After("gorm:raw").Register("instagorm:after_raw",
		postOpCb(wdB)))
}

func (wdB *wrappedDB) logError(err error) {
	if err != nil && wdB.sensor != nil && wdB.sensor.Logger() != nil {
		wdB.sensor.Logger().Error("unable to register callback, error: ", err.Error())
	}
}

func preOpCb(wdB *wrappedDB) func(db *gorm.DB) {

	return func(db *gorm.DB) {
		if db == nil || db.Statement == nil {
			wdB.sensor.Logger().Error("instagorm: preOpCb received nil db or db.Statement")
			return
		}

		ctx := db.Statement.Context

		var sqlStr string
		if db.Statement.SQL.Len() > 0 {
			sqlStr = db.Statement.SQL.String()
		}

		sp, dbKey := instana.StartSQLSpan(ctx, wdB.connDetails, sqlStr, wdB.sensor)

		if sp == nil {
			wdB.sensor.Logger().Error("instagorm: failed to start SQL span")
			return
		}

		sp.SetBaggageItem("dbKey", dbKey)

		ctx = instana.ContextWithSpan(ctx, sp)
		db.Statement.Context = ctx
	}
}

func postOpCb(wdB *wrappedDB) func(db *gorm.DB) {

	return func(db *gorm.DB) {
		if db == nil || db.Statement == nil {
			wdB.sensor.Logger().Error("instagorm: postOpCb received nil db or db.Statement")
			return
		}

		sp, ok := instana.SpanFromContext(db.Statement.Context)
		if !ok || sp == nil {
			wdB.sensor.Logger().Error("instagorm: failed to retrieve span from context")
			return
		}

		defer sp.Finish()

		dbKey := sp.BaggageItem("dbKey")

		var stmtKey string = dbKey + ".stmt"
		var errKey string = dbKey + ".error"

		if dbKey == "db" {
			stmtKey = string(ext.DBStatement)
			errKey = "error"
			if db.Dialector != nil {
				sp.SetTag(string(ext.DBType), db.Dialector.Name())
			}
		}

		if db.Statement.SQL.Len() > 0 {
			sp.SetTag(stmtKey, db.Statement.SQL.String())
		}

		if err := db.Statement.Error; err != nil {
			sp.SetTag(errKey, err.Error())
			sp.LogFields(otlog.Error(err))
		}
	}
}
