package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/Tsarbomba69-com/mammoth.server/services"
)

func TestCompareSchemas_Integration(t *testing.T) {
	t.Run("identical schemas", func(t *testing.T) {
		schemaFunc := func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		}

		source := SetupTestDB(t, "source", schemaFunc)
		target := SetupTestDB(t, "target", schemaFunc)

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 0, diff.Summary["tables_modified"])
		assert.Equal(t, []string{"users"}, diff.TablesSame)
	})

	t.Run("added table", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"])
		assert.Equal(t, 1, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 0, diff.Summary["tables_modified"])
		assert.Equal(t, "posts", diff.TablesAdded[0].Name)
	})

	t.Run("removed table", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 1, diff.Summary["tables_removed"])
		assert.Equal(t, 0, diff.Summary["tables_modified"])
		assert.Equal(t, "posts", diff.TablesRemoved[0].Name)
	})

	t.Run("modified column", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(255))")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 0, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 1, diff.Summary["tables_modified"])
		assert.Equal(t, "users", diff.TablesModified[0].Name)
		assert.Equal(t, 1, len(diff.TablesModified[0].ColumnsModified))
		assert.Equal(t, "name", diff.TablesModified[0].ColumnsModified[0].Name)
		assert.Equal(t, "TEXT", diff.TablesModified[0].ColumnsModified[0].Source.DataType)
		assert.Equal(t, "VARCHAR(255)", diff.TablesModified[0].ColumnsModified[0].Target.DataType)
	})

	t.Run("added column", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 0, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 1, diff.Summary["tables_modified"])
		assert.Equal(t, "users", diff.TablesModified[0].Name)
		assert.Equal(t, 1, len(diff.TablesModified[0].ColumnsAdded))
		assert.Equal(t, "email", diff.TablesModified[0].ColumnsAdded[0].Name)
	})

	t.Run("added index", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE INDEX idx_users_name ON users(name)")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 0, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 1, diff.Summary["tables_modified"])
		assert.Equal(t, "users", diff.TablesModified[0].Name)
		assert.Equal(t, 1, len(diff.TablesModified[0].IndexesAdded))
		assert.Equal(t, "idx_users_name", diff.TablesModified[0].IndexesAdded[0].Name)
	})

	t.Run("added foreign key", func(t *testing.T) {
		source := SetupTestDB(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
		})

		target := SetupTestDB(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES users(id))")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"]) // users table
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 1, diff.Summary["tables_modified"]) // posts table
		assert.Equal(t, "posts", diff.TablesModified[0].Name)
		assert.Equal(t, 1, len(diff.TablesModified[0].ForeignKeyInfoAdded))
	})
}
