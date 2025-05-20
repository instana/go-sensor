// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instagorm_test

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	INSERT  = "INSERT INTO `products` (`created_at`,`updated_at`,`deleted_at`,`code`,`price`) VALUES (?,?,?,?,?) RETURNING `id`"
	UPDATE  = "UPDATE `products` SET `price`=?,`updated_at`=? WHERE `products`.`deleted_at` IS NULL AND `id` = ?"
	DELETE  = "DELETE FROM `products` WHERE `products`.`id` = ?"
	SELECT  = "SELECT * FROM `products` WHERE code = ? AND `products`.`deleted_at` IS NULL ORDER BY `products`.`id` LIMIT 1"
	RAWSQL  = "SELECT * FROM products"
	DB_TYPE = "sqlite"
	ROW     = "SELECT code,price FROM `product` WHERE code = ?"
)

type product struct {
	gorm.Model
	Code  string
	Price uint
}

func TestInsertRecord(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})

		spans := recorder.GetQueuedSpans()

		span := spans[len(spans)-1]
		assert.Equal(t, 0, span.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
		require.IsType(t, instana.SDKSpanData{}, span.Data)

		data := span.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": INSERT,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func TestUpdateRecord(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})

		var p product
		db.First(&p, 1) // find product with integer primary key
		db.First(&p, "code = ?", "D42")
		db.Model(&p).Update("Price", 200)

		spans := recorder.GetQueuedSpans()

		updateSpan := spans[len(spans)-1]
		assert.Equal(t, 0, updateSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, updateSpan.Kind)
		require.IsType(t, instana.SDKSpanData{}, updateSpan.Data)

		data := updateSpan.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": UPDATE,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func TestSelectRecord(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})

		var p product
		db.First(&p, "code = ?", "D42")

		spans := recorder.GetQueuedSpans()

		selectSpan := spans[len(spans)-1]
		assert.Equal(t, 0, selectSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, selectSpan.Kind)
		require.IsType(t, instana.SDKSpanData{}, selectSpan.Data)

		data := selectSpan.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": SELECT,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func TestDeleteRecord(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})
		db.Unscoped().Delete(&product{}, 1)

		spans := recorder.GetQueuedSpans()

		deleteSpan := spans[len(spans)-1]
		assert.Equal(t, 0, deleteSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, deleteSpan.Kind)
		require.IsType(t, instana.SDKSpanData{}, deleteSpan.Data)

		data := deleteSpan.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": DELETE,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func TestRawSQL(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})
		var p product
		db.First(&p, "code = ?", "D42")

		db.Exec(RAWSQL)

		spans := recorder.GetQueuedSpans()

		rawSQLSpan := spans[5]
		assert.Equal(t, 0, rawSQLSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, rawSQLSpan.Kind)
		require.IsType(t, instana.SDKSpanData{}, rawSQLSpan.Data)

		data := rawSQLSpan.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": RAWSQL,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func TestRow(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	t.Run("Exec", func(t *testing.T) {
		dsn, tearDownFn := setupEnv(t)
		defer tearDownFn(t)

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.Statement.Context = ctx
		instagorm.Instrument(db, c, dsn)

		if err = db.AutoMigrate(&product{}); err != nil {
			panic("failed to migrate the schema")
		}
		require.NoError(t, err)

		db.Create(&product{Code: "D42", Price: 100})

		var p product
		rw := db.Table("product").Where("code = ?", "D42").Select("code", "price").Row()
		rw.Scan(&p)

		spans := recorder.GetQueuedSpans()

		rowSpan := spans[len(spans)-1]
		assert.Equal(t, 0, rowSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, rowSpan.Kind)
		require.IsType(t, instana.SDKSpanData{}, rowSpan.Data)

		data := rowSpan.Data.(instana.SDKSpanData)
		assert.Equal(t, instana.SDKSpanTags{
			Name: "sdk.database",
			Type: "exit",
			Custom: map[string]interface{}{
				"baggage": map[string]string{"dbKey": "db"},
				"tags": ot.Tags{
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  dsn,
					"db.statement": ROW,
					"db.type":      DB_TYPE,
					"peer.address": dsn,
				},
			},
		}, data.Tags)

	})
}

func setupEnv(t *testing.T) (string, func(*testing.T)) {
	db := filepath.Join(os.TempDir(), "gormtest_"+strconv.Itoa(rand.Int())+".db")

	return db, func(*testing.T) {
		err := os.Remove(db)
		if err != nil {
			t.Fatal("unable to delete the database file: ", db, ": ", err.Error())

			return
		}
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
