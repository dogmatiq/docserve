package main

import (
	"github.com/dogmatiq/docserve/persistence"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=rootpass port=25432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	models := []interface{}{
		persistence.Application{},
		persistence.Handler{},
		persistence.HandlerMessage{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			panic(err)
		}
	}

	// if err := http.ListenAndServe(":8808", &hooks.GitHubHandler{}); err != nil {
	// 	panic(err)
	// }
}
