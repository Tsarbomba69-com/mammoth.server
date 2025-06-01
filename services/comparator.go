package services

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Tsarbomba69-com/mammoth.server/models"
	"gorm.io/gorm"
)

var dialectQueries = map[string]models.QuerySet{
	"postgres": {
		Schema: `
			SELECT schema_name 
			FROM information_schema.schemata
			WHERE schema_name NOT LIKE 'pg_%'
			AND schema_name != 'information_schema'
			ORDER BY schema_name
		`,
		Table: `
			SELECT table_name as name,
			table_schema AS schema_name
			FROM information_schema.tables
			WHERE table_schema = current_schema()
			ORDER BY table_name
		`,
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
		Sequence: `
            SELECT 
                sequence_name AS name,
                sequence_schema AS schema_name,
                start_value,
                minimum_value,
                maximum_value,
                increment,
                cycle_option AS is_cyclic
            FROM information_schema.sequences
            WHERE sequence_schema = current_schema()
            ORDER BY sequence_name
        `,
		SequenceOwnership: `
            SELECT
                seq_ns.nspname AS sequence_schema,
                seq.relname AS sequence_name,
                tab_ns.nspname AS table_schema,
                tab.relname AS table_name,
                attr.attname AS column_name
            FROM pg_depend dep
            JOIN pg_class seq ON seq.oid = dep.objid
            JOIN pg_namespace seq_ns ON seq.relnamespace = seq_ns.oid
            JOIN pg_class tab ON tab.oid = dep.refobjid
            JOIN pg_namespace tab_ns ON tab.relnamespace = tab_ns.oid
            JOIN pg_attribute attr ON attr.attrelid = tab.oid AND attr.attnum = dep.refobjsubid
            WHERE dep.deptype = 'a'
            AND seq.relkind = 'S'
            AND seq_ns.nspname = current_schema()
        `,
	},
	"sqlite": {
		Schema: `
        SELECT 'main' AS schema_name
        UNION
        SELECT name AS schema_name
        FROM pragma_database_list
        WHERE name != 'main'
        ORDER BY schema_name
		`,
		Table: `
        SELECT 
            name AS name,
            CASE 
                WHEN sql LIKE '%schema%' THEN 'main' 
                ELSE 'main' 
            END AS schema_name
        FROM sqlite_master
        WHERE type = 'table'
        AND name NOT LIKE 'sqlite_%'
        ORDER BY name
		`,
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
		Sequence: `
            SELECT NULL AS name, NULL AS schema_name, NULL AS start_value,
                   NULL AS minimum_value, NULL AS maximum_value, NULL AS increment,
                   NULL AS is_cyclic, NULL AS last_value
            LIMIT 0
        `, // SQLite doesn't support sequences
		SequenceOwnership: `
            SELECT NULL AS sequence_schema, NULL AS sequence_name,
                   NULL AS table_schema, NULL AS table_name, NULL AS column_name
            LIMIT 0
        `,
	},
	"mysql": {
		Schema: `
			SELECT schema_name 
			FROM information_schema.schemata
			WHERE schema_name NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
			ORDER BY schema_name
		`,
		Table: `
			SELECT table_name,
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			ORDER BY table_name
		`,
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
		Sequence: `
            SELECT 
                sequence_name AS name,
                sequence_schema AS schema_name,
                start_value,
                minimum_value,
                maximum_value,
                increment,
                cycle_option AS is_cyclic,
                last_value
            FROM information_schema.sequences
            WHERE sequence_schema = DATABASE()
            ORDER BY sequence_name
        `,
		SequenceOwnership: `
            SELECT
                NULL AS sequence_schema,
                NULL AS sequence_name,
                NULL AS table_schema,
                NULL AS table_name,
                NULL AS column_name
            LIMIT 0
        `, // MySQL doesn't track sequence ownership like PostgreSQL
	},
}

func DumpSchema(db *gorm.DB) ([]models.Schema, error) {
	// Use channels for parallel execution
	schemasChan := make(chan []models.Schema)
	tablesChan := make(chan map[string][]struct{ Name, SchemaName string })
	columnsChan := make(chan map[string][]models.Column)
	indexesChan := make(chan map[string][]models.Index)
	fksChan := make(chan map[string][]models.ForeignKey)
	seqsChan := make(chan []models.Sequence)
	errChan := make(chan error, 6)

	// Launch goroutines for each metadata type
	go func() {
		schemas, err := getAllSchemas(db)
		if err != nil {
			errChan <- err
			return
		}
		schemasChan <- schemas
	}()

	go func() {
		tables, err := getAllTables(db)
		if err != nil {
			errChan <- err
			return
		}
		tablesChan <- tables
	}()

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

	go func() {
		seqs, err := getAllSequence(db)
		if err != nil {
			errChan <- err
			return
		}
		seqsChan <- seqs
	}()

	// Collect results
	var schemas []models.Schema
	var tables map[string][]struct{ Name, SchemaName string }
	var columnsByTable map[string][]models.Column
	var indexesByTable map[string][]models.Index
	var fksByTable map[string][]models.ForeignKey
	var sequences []models.Sequence

	for i := 0; i < 6; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case cols := <-columnsChan:
			columnsByTable = cols
		case idxs := <-indexesChan:
			indexesByTable = idxs
		case fks := <-fksChan:
			fksByTable = fks
		case ts := <-tablesChan:
			tables = ts
		case ss := <-schemasChan:
			schemas = ss
		case seqs := <-seqsChan:
			sequences = seqs
		}
	}

	// Build schemas
	for _, schema := range schemas {
		schema.Tables = make([]models.TableSchema, 0, len(tables[schema.Name]))
		schema.Sequences = sequences
		for _, table := range tables[schema.Name] {
			schema.Tables = append(schema.Tables, models.TableSchema{
				Name:        table.Name,
				SchemaName:  table.SchemaName,
				Columns:     columnsByTable[table.Name],
				Indexes:     indexesByTable[table.Name],
				ForeignKeys: fksByTable[table.Name],
			})
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func getAllSequence(db *gorm.DB) ([]models.Sequence, error) {
	var sequences []models.Sequence

	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(qs.Sequence).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query sequences: %w", err)
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			if err == nil {
				err = fmt.Errorf("rows close error: %w", closeErr)
			} else {
				err = fmt.Errorf("%v, rows close error: %w", err, closeErr)
			}
		}
	}()

	for rows.Next() {
		var seq models.Sequence
		var isCyclic string // Some dialects return string (YES/NO) for cyclic flag

		if err := rows.Scan(
			&seq.Name,
			&seq.SchemaName,
			&seq.StartValue,
			&seq.MinValue,
			&seq.MaxValue,
			&seq.Increment,
			&isCyclic,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sequence row: %w", err)
		}

		// Normalize cyclic flag
		seq.IsCyclic = isCyclic == "YES" || isCyclic == "1" || isCyclic == "true"
		sequences = append(sequences, seq)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after reading sequence rows: %w", err)
	}

	// if qs.SequenceOwnership != "" {
	// 	panic("Sequence ownership query not implemented for this dialect")
	// }

	return sequences, nil
}

func getAllSchemas(db *gorm.DB) ([]models.Schema, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}
	var schemas []string

	if err := db.Raw(qs.Schema).Scan(&schemas).Error; err != nil {
		return nil, fmt.Errorf("failed to get all schemas: %v", err)
	}

	result := make([]models.Schema, 0, len(schemas))
	for _, s := range schemas {
		result = append(result, models.Schema{
			Name: s,
		})
	}
	return result, nil
}

func getAllTables(db *gorm.DB) (map[string][]struct{ Name, SchemaName string }, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	var tables []struct {
		Name       string
		SchemaName string
	}

	if err := db.Raw(qs.Table).Scan(&tables).Error; err != nil {
		return nil, fmt.Errorf("failed to get all tables: %v", err)
	}
	result := make(map[string][]struct{ Name, SchemaName string })
	for _, c := range tables {
		result[c.SchemaName] = append(result[c.SchemaName], struct{ Name, SchemaName string }{
			Name:       c.Name,
			SchemaName: c.SchemaName,
		})
	}
	return result, nil
}

func CompareSchemas(source, target []models.Schema) models.SchemaDiff {
	var diff models.SchemaDiff
	diff.Summary = make(map[string]int)

	// Create maps for quick lookup
	sourceSchemas := make(map[string]models.Schema)
	targetSchemas := make(map[string]models.Schema)
	sourceTables := make(map[string]models.TableSchema)
	targetTables := make(map[string]models.TableSchema)
	sourceSeqs := make(map[string]models.Sequence)
	targetSeqs := make(map[string]models.Sequence)

	for _, schema := range source {
		sourceSchemas[schema.Name] = schema
		for _, table := range schema.Tables {
			sourceTables[table.Name] = table
		}

		for _, table := range schema.Sequences {
			sourceSeqs[table.Name] = table
		}
	}

	for _, schema := range target {
		targetSchemas[schema.Name] = schema
		for _, table := range schema.Tables {
			targetTables[table.Name] = table
		}

		for _, table := range schema.Sequences {
			targetSeqs[table.Name] = table
		}
	}

	// Find added and removed schemas
	for name, targetSchema := range targetSchemas {
		if _, exists := sourceSchemas[name]; !exists {
			diff.SchemasAdded = append(diff.SchemasAdded, targetSchema.Name)
		}
	}

	for name, sourceSchema := range sourceSchemas {
		if _, exists := targetSchemas[name]; !exists {
			diff.SchemasRemoved = append(diff.SchemasRemoved, sourceSchema.Name)
		} else {
			diff.SchemasSame = append(diff.SchemasSame, name)
		}
	}

	// Find added and removed tables
	for name, targetTable := range targetTables {
		if _, exists := sourceTables[name]; !exists {
			diff.TablesAdded = append(diff.TablesAdded, models.TableDiff{
				Name:            name,
				SchemaName:      targetTable.SchemaName,
				ColumnsAdded:    targetTable.Columns,
				IndexesAdded:    targetTable.Indexes,
				ForeignKeyAdded: targetTable.ForeignKeys,
			})
		}
	}

	for name, sourceTable := range sourceTables {
		if _, exists := targetTables[name]; !exists {
			diff.TablesRemoved = append(diff.TablesRemoved, models.TableDiff{
				Name:            name,
				SchemaName:      sourceTable.SchemaName,
				ColumnsAdded:    sourceTable.Columns,
				IndexesAdded:    sourceTable.Indexes,
				ForeignKeyAdded: sourceTable.ForeignKeys,
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
				len(tableDiff.ForeignKeyAdded) > 0 || len(tableDiff.ForeignKeyModified) > 0 {
				diff.TablesModified = append(diff.TablesModified, tableDiff)
			} else {
				diff.TablesSame = append(diff.TablesSame, name)
			}
		}
	}

	// Find added and removed schemas
	for name, targetSeq := range targetSeqs {
		if _, exists := sourceSeqs[name]; !exists {
			diff.SequencesAdded = append(diff.SequencesAdded, targetSeq)
		}
	}

	for name, sourceSeq := range sourceSeqs {
		if _, exists := targetSeqs[name]; !exists {
			diff.SequencesRemoved = append(diff.SequencesRemoved, sourceSeq)
		}
	}

	// Compare sequences that exist in both schemas
	for name, sourceSeq := range sourceSeqs {
		if targetSeq, exists := targetSeqs[name]; exists {
			var seqDiff = compareSequences(sourceSeq, targetSeq)
			if seqDiff.ChangedAttr != nil {
				diff.SequencesModified = append(diff.SequencesModified, seqDiff)
			} else {
				diff.SequencesSame = append(diff.SequencesSame, name)
			}
		}
	}

	// Generate summary
	diff.Summary["tables_added"] = len(diff.TablesAdded)
	diff.Summary["tables_removed"] = len(diff.TablesRemoved)
	diff.Summary["tables_modified"] = len(diff.TablesModified)
	diff.Summary["tables_same"] = len(diff.TablesSame)
	diff.Summary["schemas_added"] = len(diff.SchemasAdded)
	diff.Summary["schemas_removed"] = len(diff.SchemasRemoved)
	diff.Summary["sequences_added"] = len(diff.SequencesAdded)
	diff.Summary["sequences_removed"] = len(diff.SequencesRemoved)
	diff.Summary["sequences_modified"] = len(diff.SequencesModified)
	diff.Summary["sequences_same"] = len(diff.SequencesSame)
	return diff
}

func compareSequences(source, target models.Sequence) models.SequenceChange {
	// Convert sequences to maps for easier comparison

	if target.Name == source.Name {
		if !reflect.DeepEqual(source, target) {
			var changed []string
			if source.Increment != target.Increment {
				changed = append(changed, "increment")
			}
			if source.IsCyclic != target.IsCyclic {
				changed = append(changed, "is_cyclic")
			}
			if source.MaxValue != target.MaxValue {
				changed = append(changed, "max_value")
			}
			if source.MinValue != target.MinValue {
				changed = append(changed, "min_value")
			}

			return models.SequenceChange{
				Name:        target.Name,
				SchemaName:  target.SchemaName,
				Source:      source,
				Target:      target,
				ChangedAttr: changed,
			}
		}
	}

	return models.SequenceChange{}
}

func compareTables(source, target models.TableSchema) models.TableDiff {
	var diff models.TableDiff
	diff.Name = source.Name
	diff.SchemaName = source.SchemaName

	// Compare columns
	sourceColumns := make(map[string]models.Column)
	targetColumns := make(map[string]models.Column)

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

				diff.ColumnsModified = append(diff.ColumnsModified, models.ColumnChange{
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
	sourceIndexes := make(map[string]models.Index)
	targetIndexes := make(map[string]models.Index)

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

				diff.IndexesModified = append(diff.IndexesModified, models.IndexChange{
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

	// Compare ForeignKeys
	sourceForeignKeys := make(map[string]models.ForeignKey)
	targetForeignKeys := make(map[string]models.ForeignKey)

	for _, idx := range source.ForeignKeys {
		sourceForeignKeys[idx.Name] = idx
	}

	for _, idx := range target.ForeignKeys {
		targetForeignKeys[idx.Name] = idx
	}

	// Find added and removed ForeignKeys
	for name, idx := range targetForeignKeys {
		if _, exists := sourceForeignKeys[name]; !exists {
			diff.ForeignKeyAdded = append(diff.ForeignKeyAdded, idx)
		}
	}

	for name, idx := range sourceForeignKeys {
		if _, exists := targetForeignKeys[name]; !exists {
			diff.ForeignKeyRemoved = append(diff.ForeignKeyRemoved, idx)
		}
	}

	// Compare ForeignKeys that exist in both
	for name, sourceFk := range sourceForeignKeys {
		if targetFk, exists := targetForeignKeys[name]; exists {
			if !reflect.DeepEqual(sourceFk, targetFk) {
				var changed []string
				if !stringSlicesEqual(sourceFk.Columns, targetFk.Columns) {
					changed = append(changed, "columns")
				}
				if sourceFk.Name != targetFk.Name {
					changed = append(changed, "name")
				}
				if sourceFk.OnDelete != targetFk.OnDelete {
					changed = append(changed, "on_delete")
				}
				if sourceFk.OnUpdate != targetFk.OnUpdate {
					changed = append(changed, "on_update")
				}
				if sourceFk.ReferencedTable != targetFk.ReferencedTable {
					changed = append(changed, "referenced_table")
				}

				diff.ForeignKeyModified = append(diff.ForeignKeyModified, models.ForeignKeyChange{
					Name:        name,
					Source:      sourceFk,
					Target:      targetFk,
					ChangedAttr: changed,
				})
			} else {
				diff.ForeignKeysSame = append(diff.ForeignKeysSame, sourceFk)
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

func GetForeignKeys(db *gorm.DB, tableName string) ([]models.ForeignKey, error) {
	qs, err := getQuerySet(db)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(qs.ForeignKey, tableName).Rows()
	if err != nil {
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			if err == nil {
				err = fmt.Errorf("rows close error: %w", closeErr)
			} else {
				err = fmt.Errorf("%v, rows close error: %w", err, closeErr)
			}
		}
	}()

	constraintMap := map[string]*models.ForeignKey{}

	for rows.Next() {
		var name, column, refTable, refColumn, onUpdate, onDelete string
		if err := rows.Scan(&name, &column, &refTable, &refColumn, &onUpdate, &onDelete); err != nil {
			return nil, err
		}

		if _, exists := constraintMap[name]; !exists {
			constraintMap[name] = &models.ForeignKey{
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

	var result []models.ForeignKey
	for _, fk := range constraintMap {
		result = append(result, *fk)
	}

	return result, nil
}

func getAllColumns(db *gorm.DB) (map[string][]models.Column, error) {
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

	result := make(map[string][]models.Column)
	for _, c := range columns {
		defaultValue := ""
		if c.Default != nil {
			defaultValue = *c.Default
		}

		result[c.TableName] = append(result[c.TableName], models.Column{
			Name:       c.ColumnName,
			DataType:   c.DataType,
			IsNullable: c.IsNullable == "YES",
			IsPrimary:  c.IsPrimary,
			Default:    defaultValue,
		})
	}
	return result, nil
}

func getQuerySet(db *gorm.DB) (models.QuerySet, error) {
	dialect := db.Name()

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
		return models.QuerySet{}, fmt.Errorf("unsupported database dialect: %s", dialect)
	}

	return qs, nil
}

func getAllIndexes(db *gorm.DB) (map[string][]models.Index, error) {
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

	result := make(map[string][]models.Index)
	indexMap := make(map[string]map[string]*models.Index)

	for _, idx := range indexes {
		if _, exists := indexMap[idx.TableName]; !exists {
			indexMap[idx.TableName] = make(map[string]*models.Index)
		}

		if _, exists := indexMap[idx.TableName][idx.IndexName]; !exists {
			indexMap[idx.TableName][idx.IndexName] = &models.Index{
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

func getAllForeignKeys(db *gorm.DB) (map[string][]models.ForeignKey, error) {
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

	result := make(map[string][]models.ForeignKey)
	fkMap := make(map[string]map[string]*models.ForeignKey)

	for _, fk := range fks {
		if _, exists := fkMap[fk.TableName]; !exists {
			fkMap[fk.TableName] = make(map[string]*models.ForeignKey)
		}

		if _, exists := fkMap[fk.TableName][fk.ConstraintName]; !exists {
			fkMap[fk.TableName][fk.ConstraintName] = &models.ForeignKey{
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
