package tests

import (
	"testing"

	"github.com/Tsarbomba69-com/mammoth.server/services"
	"gorm.io/gorm"
)

func TestCodeGeneration(t *testing.T) {
	t.Run("generate migration script", func(t *testing.T) {
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
			db.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)`)
		})

		diff := services.CompareSchemas(source, target)
		migrationScript := services.Generate("postgres", diff)
		expectedUp := "CREATE TABLE \"posts\" (\n  \"id\" INTEGER,\n  \"title\" TEXT,\n  \"user_id\" INTEGER,\n  PRIMARY KEY (\"id\")\n);\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})

	t.Run("add column to existing table", func(t *testing.T) {
		source := SetupSchemaDump(t, "source_add_column", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`)
		})

		target := SetupSchemaDump(t, "target_add_column", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
		})

		diff := services.CompareSchemas(source, target)
		migrationScript := services.Generate("postgres", diff)

		expectedUp := "ALTER TABLE \"users\" ADD COLUMN \"name\" TEXT;\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})

	t.Run("drop column from existing table", func(t *testing.T) {
		source := SetupSchemaDump(t, "source_drop_column", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
		})

		target := SetupSchemaDump(t, "target_drop_column", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`)
		})

		diff := services.CompareSchemas(source, target)
		migrationScript := services.Generate("postgres", diff)

		expectedUp := "ALTER TABLE \"users\" DROP COLUMN \"name\";\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})

	t.Run("add index", func(t *testing.T) {
		source := SetupSchemaDump(t, "source_add_index", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)`)
		})

		target := SetupSchemaDump(t, "target_add_index", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)`)
			db.Exec(`CREATE UNIQUE INDEX idx_users_email ON users (email)`)
		})

		diff := services.CompareSchemas(source, target)
		migrationScript := services.Generate("postgres", diff)

		expectedUp := "CREATE UNIQUE INDEX \"idx_users_email\" ON \"users\" (\"email\");\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})

	t.Run("drop table", func(t *testing.T) {
		source := SetupSchemaDump(t, "source_drop_table", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`)
			db.Exec(`CREATE TABLE sessions (id INTEGER PRIMARY KEY)`)
		})

		target := SetupSchemaDump(t, "target_drop_table", func(db *gorm.DB) {
			db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`)
		})

		diff := services.CompareSchemas(source, target)
		migrationScript := services.Generate("postgres", diff)

		expectedUp := "DROP TABLE \"sessions\";\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})
}
