package services

import (
	"fmt"
	"strings"
)

// MigrationScript represents a generated migration script
type MigrationScript struct {
	Up   string `json:"up"`   // SQL for applying changes
	Down string `json:"down"` // SQL for reverting changes
}

// Generate creates migration scripts from schema differences
func Generate(diff SchemaDiff) MigrationScript {
	var upSQL, downSQL strings.Builder

	// Create tables first (without foreign keys)
	for _, table := range diff.TablesAdded {
		upSQL.WriteString(createTableSQL(table))
	}

	for _, table := range diff.TablesAdded {
		for _, fk := range table.ForeignKeyInfoAdded {
			upSQL.WriteString(addForeignKeySQL(table.Name, fk))
			downSQL.WriteString(dropForeignKeySQL(table.Name, fk.Name))
		}
		downSQL.WriteString(fmt.Sprintf("DROP TABLE %s;\n", quoteIdentifier(table.Name)))
	}

	// Modified tables (columns only, for now)
	for _, tableDiff := range diff.TablesModified {
		upSQL.WriteString(alterTableSQL(tableDiff))
		downSQL.WriteString(revertAlterTableSQL(tableDiff))
	}

	// Reverse: re-create removed tables (with FKs)
	for _, table := range diff.TablesRemoved {
		downSQL.WriteString(createTableSQL(table))
	}

	for _, table := range diff.TablesRemoved {
		for _, fk := range table.ForeignKeyInfoAdded {
			downSQL.WriteString(addForeignKeySQL(table.Name, fk))
			upSQL.WriteString(dropForeignKeySQL(table.Name, fk.Name))
		}
		upSQL.WriteString(fmt.Sprintf("DROP TABLE %s;\n", quoteIdentifier(table.Name)))
	}

	return MigrationScript{
		Up:   upSQL.String(),
		Down: downSQL.String(),
	}
}
func createTableSQL(tableDiff TableDiff) string {
	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdentifier(tableDiff.Name)))

	// Add columns
	for i, col := range append(tableDiff.ColumnsSame, tableDiff.ColumnsAdded...) {
		if i > 0 {
			sql.WriteString(",\n")
		}
		sql.WriteString(fmt.Sprintf("  %s %s", quoteIdentifier(col.Name), col.DataType))
		if !col.IsNullable {
			sql.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
	}

	// Add primary keys
	var pkColumns []string
	for _, col := range append(tableDiff.ColumnsSame, tableDiff.ColumnsAdded...) {
		if col.IsPrimary {
			pkColumns = append(pkColumns, quoteIdentifier(col.Name))
		}
	}
	if len(pkColumns) > 0 {
		sql.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(pkColumns, ", ")))
	}

	sql.WriteString("\n);\n")

	// Add indexes
	for _, idx := range append(tableDiff.IndexesSame, tableDiff.IndexesAdded...) {
		if !idx.IsPrimary { // Primary key already handled
			sql.WriteString(createIndexSQL(tableDiff.Name, idx))
		}
	}

	return sql.String()
}

func alterTableSQL(tableDiff TableDiff) string {
	var sql strings.Builder
	tableName := quoteIdentifier(tableDiff.Name)

	// Add columns
	for _, col := range tableDiff.ColumnsAdded {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
			tableName, quoteIdentifier(col.Name), col.DataType))
		if !col.IsNullable {
			sql.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
		sql.WriteString(";\n")
	}

	// Drop columns
	for _, col := range tableDiff.ColumnsRemoved {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n",
			tableName, quoteIdentifier(col.Name)))
	}

	// Modify columns
	for _, change := range tableDiff.ColumnsModified {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s",
			tableName, quoteIdentifier(change.Name), change.Target.DataType))
		if !change.Target.IsNullable {
			sql.WriteString(" NOT NULL")
		}
		if change.Target.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", change.Target.Default))
		}
		sql.WriteString(";\n")
	}

	// Add indexes
	for _, idx := range tableDiff.IndexesAdded {
		sql.WriteString(createIndexSQL(tableDiff.Name, idx))
	}

	// Drop indexes
	for _, idx := range tableDiff.IndexesRemoved {
		sql.WriteString(dropIndexSQL(tableDiff.Name, idx))
	}

	// Modify indexes (drop and recreate)
	for _, change := range tableDiff.IndexesModified {
		sql.WriteString(dropIndexSQL(tableDiff.Name, change.Source))
		sql.WriteString(createIndexSQL(tableDiff.Name, change.Target))
	}

	// Modify foreign key (drop and recreate)
	for _, change := range tableDiff.ForeignKeyInfoModified {
		sql.WriteString(dropForeignKeySQL(tableDiff.Name, change.Name))
		sql.WriteString(addForeignKeySQL(tableDiff.Name, change))
	}

	return sql.String()
}

func revertAlterTableSQL(tableDiff TableDiff) string {
	var sql strings.Builder
	tableName := quoteIdentifier(tableDiff.Name)

	// Revert added columns (drop them)
	for _, col := range tableDiff.ColumnsAdded {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n",
			tableName, quoteIdentifier(col.Name)))
	}

	// Revert removed columns (add them back)
	for _, col := range tableDiff.ColumnsRemoved {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
			tableName, quoteIdentifier(col.Name), col.DataType))
		if !col.IsNullable {
			sql.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
		sql.WriteString(";\n")
	}

	// Revert column modifications
	for _, change := range tableDiff.ColumnsModified {
		sql.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s",
			tableName, quoteIdentifier(change.Name), change.Source.DataType))
		if !change.Source.IsNullable {
			sql.WriteString(" NOT NULL")
		}
		if change.Source.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", change.Source.Default))
		}
		sql.WriteString(";\n")
	}

	// Revert added indexes (drop them)
	for _, idx := range tableDiff.IndexesAdded {
		sql.WriteString(dropIndexSQL(tableDiff.Name, idx))
	}

	// Revert removed indexes (add them back)
	for _, idx := range tableDiff.IndexesRemoved {
		sql.WriteString(createIndexSQL(tableDiff.Name, idx))
	}

	// Revert modified indexes
	for _, change := range tableDiff.IndexesModified {
		sql.WriteString(dropIndexSQL(tableDiff.Name, change.Target))
		sql.WriteString(createIndexSQL(tableDiff.Name, change.Source))
	}

	return sql.String()
}

func createIndexSQL(tableName string, idx IndexInfo) string {
	if idx.IsPrimary {
		return "" // Already handled in CREATE TABLE
	}

	indexType := "INDEX"
	if idx.IsUnique {
		indexType = "UNIQUE INDEX"
	}

	quotedColumns := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		quotedColumns[i] = quoteIdentifier(col)
	}

	return fmt.Sprintf("CREATE %s %s ON %s (%s);\n",
		indexType,
		quoteIdentifier(idx.Name),
		quoteIdentifier(tableName),
		strings.Join(quotedColumns, ", "))
}

func dropIndexSQL(tableName string, idx IndexInfo) string {
	if idx.IsPrimary {
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;\n",
			quoteIdentifier(tableName),
			quoteIdentifier(idx.Name))
	}
	return fmt.Sprintf("DROP INDEX %s;\n", quoteIdentifier(idx.Name))
}

func quoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func joinIdentifiers(cols []string) string {
	var parts []string
	for _, col := range cols {
		parts = append(parts, quoteIdentifier(col))
	}
	return strings.Join(parts, ", ")
}

func addForeignKeySQL(table string, fk ForeignKeyInfo) string {
	cols := joinIdentifiers(fk.Columns)
	refCols := joinIdentifiers(fk.ReferencedColumns)

	return fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s;
`, quoteIdentifier(table), quoteIdentifier(fk.Name), cols, quoteIdentifier(fk.ReferencedTable), refCols, fk.OnDelete, fk.OnUpdate)
}

func dropForeignKeySQL(table, constraint string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;\n", quoteIdentifier(table), quoteIdentifier(constraint))
}
