package services

import (
	"fmt"

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
				IsUnique: isUnique,
			})
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}
