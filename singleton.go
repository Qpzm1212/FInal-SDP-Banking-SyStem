package db

import (
	"encoding/json"
	"os"
	"sync"
)

type Database struct {
	FilePath string
}

var instance *Database
var once sync.Once

func GetDBInstance() *Database {
	once.Do(func() {
		instance = &Database{FilePath: "data/database.json"}
	})
	return instance
}

func (db *Database) SaveData(data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(db.FilePath, bytes, 0644)
}
