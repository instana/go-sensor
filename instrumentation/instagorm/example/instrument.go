package main

import (
	"database/sql/driver"

	instana "github.com/instana/go-sensor"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Wrapper struct {
	gorm.Dialector
	sensor *instana.Sensor
	driver driver.Driver
}

func DoWrap(s *instana.Sensor, d gorm.Dialector, drv driver.Driver) *Wrapper {
	instana.InstrumentSQLDriver(s, d.Name(), drv)

	return &Wrapper{
		Dialector: d,
		sensor:    s,
		driver:    drv,
	}
}

var _ gorm.Dialector = (*Wrapper)(nil)

func (w *Wrapper) Name() string {
	//fmt.Println("> Name", w.Dialector.Name())
	return w.Dialector.Name()
}

func (w *Wrapper) Initialize(db *gorm.DB) error {
	//fmt.Println("> Initialize")

	return w.Dialector.Initialize(db)
}

func (w *Wrapper) Migrator(db *gorm.DB) gorm.Migrator {
	//fmt.Println("> Migrator")
	return w.Dialector.Migrator(db)
}

func (w *Wrapper) DataTypeOf(f *schema.Field) string {
	//fmt.Println("> DataTypeOf", f.Name, w.Dialector.DataTypeOf(f))
	return w.Dialector.DataTypeOf(f)
}

func (w *Wrapper) DefaultValueOf(f *schema.Field) clause.Expression {
	//fmt.Println("> DefaultValueOf", f.Name)
	return w.Dialector.DefaultValueOf(f)
}

func (w *Wrapper) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	//fmt.Println("> BindVarTo", v)
	w.Dialector.BindVarTo(writer, stmt, v)
}

func (w *Wrapper) QuoteTo(wr clause.Writer, s string) {
	//fmt.Println("> QuoteTo", s)
	w.Dialector.QuoteTo(wr, s)
}

func (w *Wrapper) Explain(sql string, vars ...interface{}) string {
	//fmt.Println("> Explain", sql, vars, w.Dialector.Explain(sql, vars...))
	return w.Dialector.Explain(sql, vars...)
}
