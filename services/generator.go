package services

import (
	"strings"

	"github.com/Tsarbomba69-com/mammoth.server/ddl"
	"github.com/Tsarbomba69-com/mammoth.server/models"
)

// MigrationScript represents a generated migration script
type MigrationScript struct {
	Up   string `json:"up"`   // SQL for applying changes
	Down string `json:"down"` // SQL for reverting changes
}

// Generate creates migration scripts from schema differences
func Generate(dialect string, diff models.SchemaDiff) MigrationScript {
	var upSQL, downSQL strings.Builder
	var gen = ddl.NewDDL(dialect) // Change to your desired dialect

	// Create schema if it doesn't exist
	for _, schema := range diff.SchemasAdded {
		upSQL.WriteString(gen.CreateSchemaSQL(schema))
	}

	// Create sequences objects
	for _, seq := range diff.SequencesAdded {
		upSQL.WriteString(gen.CreateSequenceSQL(seq))
	}

	for _, seqDiff := range diff.SequencesModified {
		upSQL.WriteString(gen.AlterSequenceSQL(seqDiff))
		downSQL.WriteString(gen.RevertAlterSequenceSQL(seqDiff))
	}

	// Create tables first (without foreign keys)
	for _, table := range diff.TablesAdded {
		upSQL.WriteString(gen.CreateTableSQL(table))
	}

	for _, table := range diff.TablesAdded {
		for _, fk := range table.ForeignKeyAdded {
			upSQL.WriteString(gen.AddForeignKeySQL(table.SchemaName, table.Name, fk))
			downSQL.WriteString(gen.DropForeignKeySQL(table.SchemaName, table.Name, fk.Name))
		}
	}

	for _, table := range diff.TablesAdded {
		downSQL.WriteString(gen.DropTableSQL(table.SchemaName, table.Name))
	}

	for _, seq := range diff.SequencesAdded {
		downSQL.WriteString(gen.DropSequenceSQL(seq.SchemaName, seq.Name))
	}

	for _, schema := range diff.SchemasAdded {
		downSQL.WriteString(gen.DropSchemaSQL(schema))
	}

	// Modified tables (columns only, for now)
	for _, tableDiff := range diff.TablesModified {
		upSQL.WriteString(gen.AlterTableSQL(tableDiff))
		downSQL.WriteString(gen.RevertAlterTableSQL(tableDiff))
	}

	// Reverse: re-create removed tables (with FKs)
	for _, schema := range diff.SchemasRemoved {
		downSQL.WriteString(gen.CreateSchemaSQL(schema))
	}

	for _, seq := range diff.SequencesRemoved {
		downSQL.WriteString(gen.CreateSequenceSQL(seq))
	}

	for _, table := range diff.TablesRemoved {
		downSQL.WriteString(gen.CreateTableSQL(table))
	}

	for _, table := range diff.TablesRemoved {
		for _, fk := range table.ForeignKeyAdded {
			downSQL.WriteString(gen.AddForeignKeySQL(table.SchemaName, table.Name, fk))
			upSQL.WriteString(gen.DropForeignKeySQL(table.SchemaName, table.Name, fk.Name))
		}
	}

	for _, table := range diff.TablesRemoved {
		upSQL.WriteString(gen.DropTableSQL(table.SchemaName, table.Name))
	}

	for _, seq := range diff.SequencesRemoved {
		upSQL.WriteString(gen.DropSequenceSQL(seq.SchemaName, seq.Name))
	}

	for _, schema := range diff.SchemasRemoved {
		upSQL.WriteString(gen.DropSchemaSQL(schema))
	}

	return MigrationScript{
		Up:   upSQL.String(),
		Down: downSQL.String(),
	}
}
