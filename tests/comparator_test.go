package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/services"
	"github.com/Tsarbomba69-com/mammoth.server/tests/mocks"
)

func TestCompareSchemas(t *testing.T) {
	t.Run("identical schemas", func(t *testing.T) {
		schemaFunc := func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		}

		source := SetupSchemaDump(t, "source", schemaFunc)
		target := SetupSchemaDump(t, "target", schemaFunc)

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"])
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 0, diff.Summary["tables_modified"])
		assert.Equal(t, []string{"users"}, diff.TablesSame)
	})

	t.Run("added table", func(t *testing.T) {
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
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
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
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
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
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
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
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
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
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
		source := SetupSchemaDump(t, "source", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
		})

		target := SetupSchemaDump(t, "target", func(db *gorm.DB) {
			db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
			db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES users(id))")
		})

		diff := services.CompareSchemas(source, target)

		assert.Equal(t, 1, diff.Summary["tables_same"]) // users table
		assert.Equal(t, 0, diff.Summary["tables_added"])
		assert.Equal(t, 0, diff.Summary["tables_removed"])
		assert.Equal(t, 1, diff.Summary["tables_modified"]) // posts table
		assert.Equal(t, "posts", diff.TablesModified[0].Name)
		assert.Equal(t, 1, len(diff.TablesModified[0].ForeignKeyAdded))
	})

	t.Run("identical sequences", func(t *testing.T) {
		// Arrange
		source := mocks.IdenticalSourceSchema
		target := mocks.IdenticalTargetSchema
		// Act
		diff := services.CompareSchemas(source, target)
		// Assert
		// 1. Verify no sequences were added, removed, or modified
		assert.Empty(t, diff.SequencesAdded, "Expected no added sequences")
		assert.Empty(t, diff.SequencesRemoved, "Expected no removed sequences")
		assert.Empty(t, diff.SequencesModified, "Expected no modified sequences")

		// 2. Verify the identical sequence is marked as same
		assert.Len(t, diff.SequencesSame, 1, "Expected 1 sequence to be marked as same")
		assert.Contains(t, diff.SequencesSame, "seq1", "Expected seq1 to be in SequencesSame")

		// 3. Verify summary counts
		assert.Equal(t, 1, diff.Summary["sequences_same"], "Expected summary to show 1 same sequence")
		assert.Equal(t, 0, diff.Summary["sequences_added"], "Expected summary to show 0 added sequences")
		assert.Equal(t, 0, diff.Summary["sequences_removed"], "Expected summary to show 0 removed sequences")
		assert.Equal(t, 0, diff.Summary["sequences_modified"], "Expected summary to show 0 modified sequences")

		// 4. Verify no schema-level changes
		assert.Empty(t, diff.SchemasAdded, "Expected no schemas added")
		assert.Empty(t, diff.SchemasRemoved, "Expected no schemas removed")
		assert.Len(t, diff.SchemasSame, 1, "Expected 1 schema to be same")
		assert.Contains(t, diff.SchemasSame, "public", "Expected public schema to be same")
	})

	t.Run("added sequence", func(t *testing.T) {
		// Arrange
		source := []models.Schema{{
			Name:      "public",
			Sequences: []models.Sequence{},
		}}
		target := []models.Schema{{
			Name: "public",
			Sequences: []models.Sequence{{
				Name:       "seq1",
				SchemaName: "public",
				StartValue: 1,
				Increment:  1,
			}},
		}}

		// Act
		diff := services.CompareSchemas(source, target)

		// Assert
		assert.Empty(t, diff.SequencesRemoved)
		assert.Empty(t, diff.SequencesModified)
		assert.Empty(t, diff.SequencesSame)

		assert.Len(t, diff.SequencesAdded, 1)
		assert.Equal(t, "seq1", diff.SequencesAdded[0].Name)

		assert.Equal(t, 1, diff.Summary["sequences_added"])
		assert.Equal(t, 0, diff.Summary["sequences_removed"])
		assert.Equal(t, 0, diff.Summary["sequences_modified"])
		assert.Equal(t, 0, diff.Summary["sequences_same"])
	})
}
