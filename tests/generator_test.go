package tests

import (
	"testing"

	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/services"
	"github.com/stretchr/testify/assert"
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
		expectedUp := "CREATE TABLE \"main\".\"posts\" (\n  \"id\" INTEGER,\n  \"title\" TEXT,\n  \"user_id\" INTEGER,\n  PRIMARY KEY (\"id\")\n);\n"

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

		expectedUp := "ALTER TABLE \"main\".\"users\" ADD COLUMN \"name\" TEXT;\n"

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

		expectedUp := "ALTER TABLE \"main\".\"users\" DROP COLUMN \"name\";\n"

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

		expectedUp := "CREATE UNIQUE INDEX \"idx_users_email\" ON \"main\".\"users\" (\"email\");\n"

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

		expectedUp := "DROP TABLE \"main\".\"sessions\";\n"

		if migrationScript.Up != expectedUp {
			t.Errorf("expected migration script:\n%s\nbut got:\n%s", expectedUp, migrationScript.Up)
		}
	})
}

func TestGenerate_AddedSequence(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		diff     models.SchemaDiff
		expected services.MigrationScript
	}{
		{
			name:    "simple sequence addition",
			dialect: "postgres",
			diff: models.SchemaDiff{
				SequencesAdded: []models.Sequence{
					{
						Name:       "seq_user_id",
						SchemaName: "public",
						StartValue: 1,
						Increment:  1,
						MinValue:   1,
						MaxValue:   1000,
						IsCyclic:   false,
					},
				},
				Summary: map[string]int{
					"sequences_added": 1,
				},
			},
			expected: services.MigrationScript{
				Up:   "CREATE SEQUENCE \"public\".\"seq_user_id\" INCREMENT BY 1 MINVALUE 1 MAXVALUE 1000 START WITH 1 NO CYCLE;\n",
				Down: "DROP SEQUENCE IF EXISTS \"public\".\"seq_user_id\";\n",
			},
		},
		{
			name:    "sequence with ownership",
			dialect: "postgres",
			diff: models.SchemaDiff{
				SequencesAdded: []models.Sequence{
					{
						Name:          "seq_order_id",
						SchemaName:    "public",
						StartValue:    100,
						Increment:     2,
						OwnedByTable:  "orders",
						OwnedByColumn: "id",
					},
				},
				Summary: map[string]int{
					"sequences_added": 1,
				},
			},
			expected: services.MigrationScript{
				Up: `CREATE SEQUENCE "public"."seq_order_id" INCREMENT BY 2 START WITH 100 NO CYCLE;
 ALTER SEQUENCE "public"."seq_order_id" OWNED BY "orders"."id";
`,
				Down: `DROP SEQUENCE IF EXISTS "public"."seq_order_id";
`,
			},
		},
		{
			name:    "cyclic sequence",
			dialect: "postgres",
			diff: models.SchemaDiff{
				SequencesAdded: []models.Sequence{
					{
						Name:       "seq_cycle",
						SchemaName: "public",
						StartValue: 1,
						Increment:  1,
						IsCyclic:   true,
					},
				},
				Summary: map[string]int{
					"sequences_added": 1,
				},
			},
			expected: services.MigrationScript{
				Up:   "CREATE SEQUENCE \"public\".\"seq_cycle\" INCREMENT BY 1 START WITH 1 CYCLE;\n",
				Down: "DROP SEQUENCE IF EXISTS \"public\".\"seq_cycle\";\n",
			},
		},
		{
			name:    "multiple sequences added",
			dialect: "postgres",
			diff: models.SchemaDiff{
				SequencesAdded: []models.Sequence{
					{
						Name:       "seq_one",
						SchemaName: "public",
						StartValue: 1,
						Increment:  1,
					},
					{
						Name:          "seq_two",
						SchemaName:    "app",
						StartValue:    100,
						Increment:     10,
						OwnedByTable:  "app.users",
						OwnedByColumn: "user_id",
					},
				},
				Summary: map[string]int{
					"sequences_added": 2,
				},
			},
			expected: services.MigrationScript{
				Up: `CREATE SEQUENCE "public"."seq_one" INCREMENT BY 1 START WITH 1 NO CYCLE;
CREATE SEQUENCE "app"."seq_two" INCREMENT BY 10 START WITH 100 NO CYCLE;
 ALTER SEQUENCE "app"."seq_two" OWNED BY "app.users"."user_id";
`,
				Down: `DROP SEQUENCE IF EXISTS "public"."seq_one";
DROP SEQUENCE IF EXISTS "app"."seq_two";
`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := services.Generate(tt.dialect, tt.diff)

			// Assert
			assert.Equal(t, tt.expected.Up, result.Up, "Up migration mismatch")
			assert.Equal(t, tt.expected.Down, result.Down, "Down migration mismatch")
		})
	}
}
