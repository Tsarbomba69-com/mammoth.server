package services

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

type TableSchema struct {
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
	Indexes []IndexInfo  `json:"indexes"`
}

type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimary  bool   `json:"is_primary"`
	Default    string `json:"default"`
}

type IndexInfo struct {
	Name      string   `json:"name"`
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
}

type SchemaDiff struct {
	TablesAdded    []TableDiff    `json:"tables_added"`
	TablesRemoved  []TableDiff    `json:"tables_removed"`
	TablesModified []TableDiff    `json:"tables_modified"`
	TablesSame     []string       `json:"tables_same"`
	Summary        map[string]int `json:"summary"`
}

type TableDiff struct {
	Name            string         `json:"table_name"`
	ColumnsAdded    []ColumnInfo   `json:"columns_added"`
	ColumnsRemoved  []ColumnInfo   `json:"columns_removed"`
	ColumnsModified []ColumnChange `json:"columns_modified"`
	ColumnsSame     []ColumnInfo   `json:"columns_same"`
	IndexesAdded    []IndexInfo    `json:"indexes_added"`
	IndexesRemoved  []IndexInfo    `json:"indexes_removed"`
	IndexesModified []IndexChange  `json:"indexes_modified"`
	IndexesSame     []IndexInfo    `json:"indexes_same"`
}

type ColumnChange struct {
	Name        string     `json:"name"`
	Source      ColumnInfo `json:"source"`
	Target      ColumnInfo `json:"target"`
	ChangedAttr []string   `json:"changed_attributes"`
}

type IndexChange struct {
	Name        string    `json:"name"`
	Source      IndexInfo `json:"source"`
	Target      IndexInfo `json:"target"`
	ChangedAttr []string  `json:"changed_attributes"`
}

// TODO: Deal with constraints
func DumpSchemaAST(db *gorm.DB) ([]TableSchema, error) {
	var schemas []TableSchema

	// Get all tables in the database
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %v", err)
	}

	for _, table := range tables {
		var schema TableSchema
		schema.Name = table

		columns, err := db.Migrator().ColumnTypes(table)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %v", table, err)
		}

		for _, column := range columns {
			isNull, _ := column.Nullable()
			isPrimary, _ := column.PrimaryKey()
			schema.Columns = append(schema.Columns, ColumnInfo{
				Name:       column.Name(),
				DataType:   column.DatabaseTypeName(),
				IsNullable: isNull,
				IsPrimary:  isPrimary,
			})
		}

		// Get indexes for the table
		indexes, err := db.Migrator().GetIndexes(table)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %v", table, err)
		}

		for _, index := range indexes {
			isUnique, _ := index.Unique()
			schema.Indexes = append(schema.Indexes, IndexInfo{
				Name:     index.Name(),
				Columns:  index.Columns(),
				IsUnique: isUnique,
			})
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func CompareSchemas(source, target []TableSchema) SchemaDiff {
	var diff SchemaDiff
	diff.Summary = make(map[string]int)

	// Create maps for quick lookup
	sourceTables := make(map[string]TableSchema)
	targetTables := make(map[string]TableSchema)

	for _, table := range source {
		sourceTables[table.Name] = table
	}

	for _, table := range target {
		targetTables[table.Name] = table
	}

	// Find added and removed tables
	for name, targetTable := range targetTables {
		if _, exists := sourceTables[name]; !exists {
			diff.TablesAdded = append(diff.TablesAdded, TableDiff{
				Name:         name,
				ColumnsAdded: targetTable.Columns,
				IndexesAdded: targetTable.Indexes,
			})
		}
	}

	for name, sourceTable := range sourceTables {
		if _, exists := targetTables[name]; !exists {
			diff.TablesRemoved = append(diff.TablesRemoved, TableDiff{
				Name:         name,
				ColumnsAdded: sourceTable.Columns,
				IndexesAdded: sourceTable.Indexes,
			})
		}
	}

	// Compare tables that exist in both schemas
	for name, sourceTable := range sourceTables {
		if targetTable, exists := targetTables[name]; exists {
			tableDiff := compareTables(sourceTable, targetTable)
			if len(tableDiff.ColumnsAdded) > 0 || len(tableDiff.ColumnsRemoved) > 0 ||
				len(tableDiff.ColumnsModified) > 0 || len(tableDiff.IndexesAdded) > 0 ||
				len(tableDiff.IndexesRemoved) > 0 || len(tableDiff.IndexesModified) > 0 {
				diff.TablesModified = append(diff.TablesModified, tableDiff)
			} else {
				diff.TablesSame = append(diff.TablesSame, name)
			}
		}
	}

	// Generate summary
	diff.Summary["tables_added"] = len(diff.TablesAdded)
	diff.Summary["tables_removed"] = len(diff.TablesRemoved)
	diff.Summary["tables_modified"] = len(diff.TablesModified)
	diff.Summary["tables_same"] = len(diff.TablesSame)
	return diff
}

func compareTables(source, target TableSchema) TableDiff {
	var diff TableDiff
	diff.Name = source.Name

	// Compare columns
	sourceColumns := make(map[string]ColumnInfo)
	targetColumns := make(map[string]ColumnInfo)

	for _, col := range source.Columns {
		sourceColumns[col.Name] = col
	}

	for _, col := range target.Columns {
		targetColumns[col.Name] = col
	}

	// Find added and removed columns
	for name, col := range targetColumns {
		if _, exists := sourceColumns[name]; !exists {
			diff.ColumnsAdded = append(diff.ColumnsAdded, col)
		}
	}

	for name, col := range sourceColumns {
		if _, exists := targetColumns[name]; !exists {
			diff.ColumnsRemoved = append(diff.ColumnsRemoved, col)
		}
	}

	// Compare columns that exist in both
	for name, sourceCol := range sourceColumns {
		if targetCol, exists := targetColumns[name]; exists {
			if !reflect.DeepEqual(sourceCol, targetCol) {
				var changed []string
				if sourceCol.DataType != targetCol.DataType {
					changed = append(changed, "data_type")
				}
				if sourceCol.IsNullable != targetCol.IsNullable {
					changed = append(changed, "is_nullable")
				}
				if sourceCol.IsPrimary != targetCol.IsPrimary {
					changed = append(changed, "is_primary")
				}
				if sourceCol.Default != targetCol.Default {
					changed = append(changed, "default")
				}

				diff.ColumnsModified = append(diff.ColumnsModified, ColumnChange{
					Name:        name,
					Source:      sourceCol,
					Target:      targetCol,
					ChangedAttr: changed,
				})
			} else {
				diff.ColumnsSame = append(diff.ColumnsSame, sourceCol)
			}
		}
	}

	// Compare indexes
	sourceIndexes := make(map[string]IndexInfo)
	targetIndexes := make(map[string]IndexInfo)

	for _, idx := range source.Indexes {
		sourceIndexes[idx.Name] = idx
	}

	for _, idx := range target.Indexes {
		targetIndexes[idx.Name] = idx
	}

	// Find added and removed indexes
	for name, idx := range targetIndexes {
		if _, exists := sourceIndexes[name]; !exists {
			diff.IndexesAdded = append(diff.IndexesAdded, idx)
		}
	}

	for name, idx := range sourceIndexes {
		if _, exists := targetIndexes[name]; !exists {
			diff.IndexesRemoved = append(diff.IndexesRemoved, idx)
		}
	}

	// Compare indexes that exist in both
	for name, sourceIdx := range sourceIndexes {
		if targetIdx, exists := targetIndexes[name]; exists {
			if !reflect.DeepEqual(sourceIdx, targetIdx) {
				var changed []string
				if !stringSlicesEqual(sourceIdx.Columns, targetIdx.Columns) {
					changed = append(changed, "columns")
				}
				if sourceIdx.IsUnique != targetIdx.IsUnique {
					changed = append(changed, "is_unique")
				}
				if sourceIdx.IsPrimary != targetIdx.IsPrimary {
					changed = append(changed, "is_primary")
				}

				diff.IndexesModified = append(diff.IndexesModified, IndexChange{
					Name:        name,
					Source:      sourceIdx,
					Target:      targetIdx,
					ChangedAttr: changed,
				})
			} else {
				diff.IndexesSame = append(diff.IndexesSame, sourceIdx)
			}
		}
	}

	return diff
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
