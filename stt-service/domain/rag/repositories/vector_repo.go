package repositories

import (
	"fmt"
	"sync"

	"github.com/philippgille/chromem-go"
)

type VectorRepository interface {
	GetDB() *chromem.DB
}

type vectorRepository struct {
	db *chromem.DB
}

var (
	instance *vectorRepository
	once     sync.Once
)

func NewVectorRepository() VectorRepository {
	once.Do(func() {
		// Using in-memory DB for now
		db := chromem.NewDB()

		// Create a default collection if it doesn't exist
		_, err := db.CreateCollection("default", nil, nil)
		if err != nil {
			fmt.Printf("Error creating default collection: %v\n", err)
		}

		instance = &vectorRepository{
			db: db,
		}
	})
	return instance
}

func (r *vectorRepository) GetDB() *chromem.DB {
	return r.db
}
