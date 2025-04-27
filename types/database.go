package types

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
	Name                   string             `json:"table_name"`
	ColumnsAdded           []ColumnInfo       `json:"columns_added"`
	ColumnsRemoved         []ColumnInfo       `json:"columns_removed"`
	ColumnsModified        []ColumnChange     `json:"columns_modified"`
	ColumnsSame            []ColumnInfo       `json:"columns_same"`
	IndexesAdded           []IndexInfo        `json:"indexes_added"`
	IndexesRemoved         []IndexInfo        `json:"indexes_removed"`
	IndexesModified        []IndexChange      `json:"indexes_modified"`
	IndexesSame            []IndexInfo        `json:"indexes_same"`
	ForeignKeyInfoAdded    []ForeignKeyInfo   `json:"foreign_key_info_added"`
	ForeignKeyInfoModified []ForeignKeyChange `json:"foreign_key_info_modified"`
	ForeignKeyInfoRemoved  []ForeignKeyInfo   `json:"foreign_key_info_removed"`
	ForeignKeysInfoSame    []ForeignKeyInfo   `json:"foreign_key_info_same"`
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

type ForeignKeyChange struct {
	Name        string         `json:"name"`
	Source      ForeignKeyInfo `json:"source"`
	Target      ForeignKeyInfo `json:"target"`
	ChangedAttr []string       `json:"changed_attributes"`
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
