// (c) Copyright IBM Corp. 2023

package main

import (
	"fmt"
	"time"

	instana "github.com/instana/go-sensor"
	sqlite3 "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	s := instana.NewSensor("my-gorm-service")
	d := sqlite.Open("test.db")
	wrapped := DoWrap(s, d, &sqlite3.SQLiteDriver{})

	db, err := gorm.Open(wrapped, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	time.Sleep(time.Second * 2)

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "D42", Price: 100})

	// Read
	var product Product
	// db.First(&product, 1) // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	fmt.Println("Data:", product)

	// Update - update product's price to 200
	// db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	// db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	// db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	db.Unscoped().Delete(&product, 1).Commit()

	// time.Sleep(time.Second * 10)
	// fmt.Println("fim")
	ch := make(chan int)
	<-ch
}
