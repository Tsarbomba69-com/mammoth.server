package tests

import (
	"fmt"
	"testing"

	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/services"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func SetupSchemaDump(t *testing.T, dbName string, schemaFunc func(*gorm.DB)) []models.Schema {
	// Create in-memory SQLite database with proper name substitution
	db := SetupDB(t, dbName, schemaFunc)

	// Dump the schema
	schemas, err := services.DumpSchemaAST(db)
	if err != nil {
		t.Fatalf("failed to dump schema: %v", err)
	}
	return schemas
}

func SetupDB(t *testing.T, dbName string, schemaFunc func(*gorm.DB)) *gorm.DB {
	// Create in-memory SQLite database with proper name substitution
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	// Apply schema changes
	schemaFunc(db)
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			err = sqlDB.Close() // This will destroy the in-memory database
			if err != nil {
				t.Fatalf("failed to close database: %v", err)
			}
		}
	})
	return db
}
