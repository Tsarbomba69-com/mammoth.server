package services

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type TableSchema struct {
	Name        string           `json:"name"`
	Columns     []ColumnInfo     `json:"columns"`
	Indexes     []IndexInfo      `json:"indexes"`
	ForeignKeys []ForeignKeyInfo `json:"foreign_keys"`
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
	Name                   string           `json:"table_name"`
	ColumnsAdded           []ColumnInfo     `json:"columns_added"`
	ColumnsRemoved         []ColumnInfo     `json:"columns_removed"`
	ColumnsModified        []ColumnChange   `json:"columns_modified"`
	ColumnsSame            []ColumnInfo     `json:"columns_same"`
	IndexesAdded           []IndexInfo      `json:"indexes_added"`
	IndexesRemoved         []IndexInfo      `json:"indexes_removed"`
	IndexesModified        []IndexChange    `json:"indexes_modified"`
	IndexesSame            []IndexInfo      `json:"indexes_same"`
	ForeignKeyInfoAdded    []ForeignKeyInfo `json:"foreign_key_info_added"`
	ForeignKeyInfoModified []ForeignKeyInfo `json:"foreign_key_info_modified"`
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

type ForeignKeyInfo struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
	OnUpdate          string
}

type QuerySet struct {
	Column     string
	Index      string
	ForeignKey string
}

var dialectQueries = map[string]QuerySet{
	"postgres": {
		Column: `
			SELECT 
				c.table_name,
				c.column_name,
				c.data_type,
				c.is_nullable,
				EXISTS (
					SELECT 1 FROM information_schema.key_column_usage k
					WHERE k.table_name = c.table_name 
					AND k.column_name = c.column_name
					AND k.constraint_name IN (
						SELECT constraint_name 
						FROM information_schema.table_constraints 
						WHERE constraint_type = 'PRIMARY KEY'
					)
				) AS is_primary,
				c.column_default AS default_value
			FROM information_schema.columns c
			WHERE c.table_schema = current_schema()
			ORDER BY c.table_name, c.ordinal_position
		`,
		Index: `
			SELECT
				t.relname AS table_name,
				i.relname AS index_name,
				a.attname AS column_name,
				idx.indisunique AS is_unique,
				idx.indisprimary AS is_primary
			FROM
				pg_class t,
				pg_class i,
				pg_index idx,
				pg_attribute a
			WHERE
				t.oid = idx.indrelid
				AND i.oid = idx.indexrelid
				AND a.attrelid = t.oid
				AND a.attnum = ANY(idx.indkey)
				AND t.relkind = 'r'
				AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = current_schema())
			ORDER BY
				t.relname,
				i.relname,
				array_position(idx.indkey, a.attnum)
		`,
		ForeignKey: `
			SELECT
				tc.table_name,
				tc.constraint_name,
				kcu.column_name,
				ccu.table_name AS foreign_table,
				ccu.column_name AS foreign_column,
				rc.delete_rule AS on_delete,
				rc.update_rule AS on_update
			FROM
				information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
					AND tc.table_name = kcu.table_name
				JOIN information_schema.constraint_column_usage ccu
					ON ccu.constraint_name = tc.constraint_name
					AND ccu.table_schema = tc.table_schema
				JOIN information_schema.referential_constraints rc
					ON rc.constraint_name = tc.constraint_name
					AND rc.constraint_schema = tc.table_schema
			WHERE
				tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_schema = current_schema()
			ORDER BY
				tc.table_name,
				tc.constraint_name,
				kcu.ordinal_position
		`,
	},
	"sqlite": {
		Column: `
			SELECT 
				m.name AS table_name,
				p.name AS column_name,
				p.type AS data_type,
				CASE WHEN p."notnull" = 0 THEN 'YES' ELSE 'NO' END AS is_nullable,
				p.pk > 0 AS is_primary,
				p.dflt_value AS default_value
			FROM sqlite_master m
			JOIN pragma_table_info(m.name) p
			WHERE m.type = 'table'
			ORDER BY m.name, p.cid
		`,
		Index: `
			SELECT
				m.name AS table_name,
				il.name AS index_name,
				ii.name AS column_name,
				il."unique" AS is_unique,
				il.origin = 'pk' AS is_primary
			FROM sqlite_master m
			JOIN pragma_index_list(m.name) il
			JOIN pragma_index_info(il.name) ii
			WHERE m.type = 'table'
			ORDER BY m.name, il.name, ii.seqno
		`,
		ForeignKey: `
			SELECT
				m.name AS table_name,
				fk.id AS constraint_name,
				fk."from" AS column_name,
				fk."table" AS foreign_table,
				fk."to" AS foreign_column,
				fk.on_delete AS on_delete,
				fk.on_update AS on_update
			FROM sqlite_master m
			JOIN pragma_foreign_key_list(m.name) fk
			WHERE m.type = 'table'
			ORDER BY m.name, fk.id, fk.seq
		`,
	},
	"mysql": {
		Column: `
			SELECT 
				table_name,
				column_name,
				data_type,
				is_nullable,
				column_key = 'PRI' AS is_primary,
				column_default AS default_value
			FROM information_schema.columns
			WHERE table_schema = DATABASE()
			ORDER BY table_name, ordinal_position
		`,
		Index: `
			SELECT
				table_name,
				index_name,
				column_name,
				non_unique = 0 AS is_unique,
				index_name = 'PRIMARY' AS is_primary
			FROM information_schema.statistics
			WHERE table_schema = DATABASE()
			ORDER BY table_name, index_name, seq_in_index
		`,
		ForeignKey: `
			SELECT
				table_name,
				constraint_name,
				column_name,
				referenced_table_name AS foreign_table_name,
				referenced_column_name AS foreign_column_name,
				delete_rule AS on_delete,
				update_rule AS on_update
			FROM information_schema.key_column_usage
			WHERE table_schema = DATABASE()
			AND referenced_table_name IS NOT NULL
			ORDER BY table_name, constraint_name, ordinal_position
		`,
	},
}

func DumpSchemaAST(db *gorm.DB) ([]TableSchema, error) {
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	// Use channels for parallel execution
	columnsChan := make(chan map[string][]ColumnInfo)
	indexesChan := make(chan map[string][]IndexInfo)
	fksChan := make(chan map[string][]ForeignKeyInfo)
	errChan := make(chan error, 3)

	// Launch goroutines for each metadata type
	go func() {
		cols, err := getAllColumns(db)
		if err != nil {
			errChan <- err
			return
		}
		columnsChan <- cols
	}()

	go func() {
		idxs, err := getAllIndexes(db)
		if err != nil {
			errChan <- err
			return
		}
		indexesChan <- idxs
	}()

	go func() {
		fks, err := getAllForeignKeys(db)
		if err != nil {
			errChan <- err
			return
		}
		fksChan <- fks
	}()

	// Collect results
	var columnsByTable map[string][]ColumnInfo
	var indexesByTable map[string][]IndexInfo
	var fksByTable map[string][]ForeignKeyInfo

	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case cols := <-columnsChan:
			columnsByTable = cols
		case idxs := <-indexesChan:
			indexesByTable = idxs
		case fks := <-fksChan:
			fksByTable = fks
		}
	}

	// Build schemas
	schemas := make([]TableSchema, 0, len(tables))
	for _, table := range tables {
		schemas = append(schemas, TableSchema{
			Name:        table,
			Columns:     columnsByTable[table],
			Indexes:     indexesByTable[table],
			ForeignKeys: fksByTable[table],
		})
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
				Name:                name,
				ColumnsAdded:        targetTable.Columns,
				IndexesAdded:        targetTable.Indexes,
				ForeignKeyInfoAdded: targetTable.ForeignKeys,
			})
		}
	}

	for name, sourceTable := range sourceTables {
		if _, exists := targetTables[name]; !exists {
			diff.TablesRemoved = append(diff.TablesRemoved, TableDiff{
				Name:                name,
				ColumnsAdded:        sourceTable.Columns,
				IndexesAdded:        sourceTable.Indexes,
				ForeignKeyInfoAdded: sourceTable.ForeignKeys,
			})
		}
	}

	// Compare tables that exist in both schemas
	for name, sourceTable := range sourceTables {
		if targetTable, exists := targetTables[name]; exists {
			tableDiff := compareTables(sourceTable, targetTable)
			if len(tableDiff.ColumnsAdded) > 0 || len(tableDiff.ColumnsRemoved) > 0 ||
				len(tableDiff.ColumnsModified) > 0 || len(tableDiff.IndexesAdded) > 0 ||
				len(tableDiff.IndexesRemoved) > 0 || len(tableDiff.IndexesModified) > 0 ||
				len(tableDiff.ForeignKeyInfoAdded) > 0 || len(tableDiff.ForeignKeyInfoModified) > 0 {
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

// TODO: Implement the foreign key comparison logic in compareTables function
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

func GetForeignKeys(db *gorm.DB, tableName string) ([]ForeignKeyInfo, error) {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS referenced_table,
			ccu.column_name AS referenced_column,
			rc.update_rule,
			rc.delete_rule
		FROM
			information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_name = kcu.table_name
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE
			tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = $1
		ORDER BY
			tc.constraint_name, kcu.ordinal_position;
	`

	rows, err := db.Raw(query, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	constraintMap := map[string]*ForeignKeyInfo{}

	for rows.Next() {
		var name, column, refTable, refColumn, onUpdate, onDelete string
		if err := rows.Scan(&name, &column, &refTable, &refColumn, &onUpdate, &onDelete); err != nil {
			return nil, err
		}

		if _, exists := constraintMap[name]; !exists {
			constraintMap[name] = &ForeignKeyInfo{
				Name:            name,
				ReferencedTable: refTable,
				OnUpdate:        onUpdate,
				OnDelete:        onDelete,
			}
		}
		info := constraintMap[name]
		info.Columns = append(info.Columns, column)
		info.ReferencedColumns = append(info.ReferencedColumns, refColumn)
	}

	var result []ForeignKeyInfo
	for _, fk := range constraintMap {
		result = append(result, *fk)
	}

	return result, nil
}

func getAllColumns(db *gorm.DB) (map[string][]ColumnInfo, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	var columns []struct {
		TableName  string
		ColumnName string
		DataType   string
		IsNullable string
		IsPrimary  bool
		Default    *string
	}

	if err := db.Raw(qs.Column).Scan(&columns).Error; err != nil {
		return nil, fmt.Errorf("failed to get all columns: %v", err)
	}

	result := make(map[string][]ColumnInfo)
	for _, c := range columns {
		defaultValue := ""
		if c.Default != nil {
			defaultValue = *c.Default
		}

		result[c.TableName] = append(result[c.TableName], ColumnInfo{
			Name:       c.ColumnName,
			DataType:   c.DataType,
			IsNullable: c.IsNullable == "YES",
			IsPrimary:  c.IsPrimary,
			Default:    defaultValue,
		})
	}
	return result, nil
}

func getQuerySet(db *gorm.DB) (QuerySet, error) {
	dialect := db.Dialector.Name()

	// Normalize dialect names
	switch {
	case strings.Contains(dialect, "postgres"):
		dialect = "postgres"
	case strings.Contains(dialect, "sqlite"):
		dialect = "sqlite"
	case strings.Contains(dialect, "mysql"):
		dialect = "mysql"
	}

	qs, ok := dialectQueries[dialect]
	if !ok {
		return QuerySet{}, fmt.Errorf("unsupported database dialect: %s", dialect)
	}

	return qs, nil
}

func getAllIndexes(db *gorm.DB) (map[string][]IndexInfo, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	var indexes []struct {
		TableName  string
		IndexName  string
		ColumnName string
		IsUnique   bool
		IsPrimary  bool
	}

	if err := db.Raw(qs.Index).Scan(&indexes).Error; err != nil {
		return nil, fmt.Errorf("failed to get all indexes: %v", err)
	}

	result := make(map[string][]IndexInfo)
	indexMap := make(map[string]map[string]*IndexInfo)

	for _, idx := range indexes {
		if _, exists := indexMap[idx.TableName]; !exists {
			indexMap[idx.TableName] = make(map[string]*IndexInfo)
		}

		if _, exists := indexMap[idx.TableName][idx.IndexName]; !exists {
			indexMap[idx.TableName][idx.IndexName] = &IndexInfo{
				Name:      idx.IndexName,
				IsUnique:  idx.IsUnique,
				IsPrimary: idx.IsPrimary,
			}
		}

		indexMap[idx.TableName][idx.IndexName].Columns = append(
			indexMap[idx.TableName][idx.IndexName].Columns,
			idx.ColumnName,
		)
	}

	for tableName, indexes := range indexMap {
		for _, index := range indexes {
			result[tableName] = append(result[tableName], *index)
		}
	}

	return result, nil
}

func getAllForeignKeys(db *gorm.DB) (map[string][]ForeignKeyInfo, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	var fks []struct {
		TableName      string
		ConstraintName string
		ColumnName     string
		ForeignTable   string
		ForeignColumn  string
		OnDelete       string
		OnUpdate       string
	}

	if err := db.Raw(qs.ForeignKey).Scan(&fks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all foreign keys: %v", err)
	}

	result := make(map[string][]ForeignKeyInfo)
	fkMap := make(map[string]map[string]*ForeignKeyInfo)

	for _, fk := range fks {
		if _, exists := fkMap[fk.TableName]; !exists {
			fkMap[fk.TableName] = make(map[string]*ForeignKeyInfo)
		}

		if _, exists := fkMap[fk.TableName][fk.ConstraintName]; !exists {
			fkMap[fk.TableName][fk.ConstraintName] = &ForeignKeyInfo{
				Name:            fk.ConstraintName,
				ReferencedTable: fk.ForeignTable,
				OnDelete:        fk.OnDelete,
				OnUpdate:        fk.OnUpdate,
			}
		}

		if !contains(fkMap[fk.TableName][fk.ConstraintName].Columns, fk.ColumnName) {
			fkMap[fk.TableName][fk.ConstraintName].Columns = append(
				fkMap[fk.TableName][fk.ConstraintName].Columns,
				fk.ColumnName,
			)
		}

		if !contains(fkMap[fk.TableName][fk.ConstraintName].ReferencedColumns, fk.ForeignColumn) {
			fkMap[fk.TableName][fk.ConstraintName].ReferencedColumns = append(
				fkMap[fk.TableName][fk.ConstraintName].ReferencedColumns,
				fk.ForeignColumn,
			)
		}
	}

	for tableName, constraints := range fkMap {
		for _, constraint := range constraints {
			result[tableName] = append(result[tableName], *constraint)
		}
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
