package models

type Schema struct {
	Name      string        `json:"name"`
	Tables    []TableSchema `json:"tables"`
	Sequences []Sequence    `json:"sequences"`
}

type TableSchema struct {
	Name        string       `json:"name"`
	SchemaName  string       `json:"schema_name"`
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
}

type Column struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimary  bool   `json:"is_primary"`
	Default    string `json:"default"`
}

type Index struct {
	Name      string   `json:"name"`
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
}

type Sequence struct {
	Name          string
	SchemaName    string
	StartValue    int64
	MinValue      int64
	MaxValue      int64
	Increment     int64
	IsCyclic      bool
	OwnedByTable  string // Only populated if the sequence is owned by a table column
	OwnedByColumn string
}

type SequenceChange struct {
	Name        string
	SchemaName  string
	Source      Sequence
	Target      Sequence
	ChangedAttr []string `json:"changed_attributes"`
}

type SchemaDiff struct {
	SchemasAdded      []string         `json:"schemas_added"`
	SchemasSame       []string         `json:"schemas_same"`
	SchemasRemoved    []string         `json:"schemas_removed"`
	TablesAdded       []TableDiff      `json:"tables_added"`
	TablesRemoved     []TableDiff      `json:"tables_removed"`
	TablesModified    []TableDiff      `json:"tables_modified"`
	TablesSame        []string         `json:"tables_same"`
	SequencesAdded    []Sequence       `json:"sequences_added"`
	SequencesSame     []string         `json:"sequences_same"`
	SequencesRemoved  []Sequence       `json:"sequences_removed"`
	SequencesModified []SequenceChange `json:"sequences_modified"`
	Summary           map[string]int   `json:"summary"`
}

type TableDiff struct {
	Name               string             `json:"table_name"`
	SchemaName         string             `json:"schema_name"`
	ColumnsAdded       []Column           `json:"columns_added"`
	ColumnsRemoved     []Column           `json:"columns_removed"`
	ColumnsModified    []ColumnChange     `json:"columns_modified"`
	ColumnsSame        []Column           `json:"columns_same"`
	IndexesAdded       []Index            `json:"indexes_added"`
	IndexesRemoved     []Index            `json:"indexes_removed"`
	IndexesModified    []IndexChange      `json:"indexes_modified"`
	IndexesSame        []Index            `json:"indexes_same"`
	ForeignKeyAdded    []ForeignKey       `json:"foreign_key_added"`
	ForeignKeyModified []ForeignKeyChange `json:"foreign_key_modified"`
	ForeignKeyRemoved  []ForeignKey       `json:"foreign_key_removed"`
	ForeignKeysSame    []ForeignKey       `json:"foreign_key_same"`
}

type ColumnChange struct {
	Name        string   `json:"name"`
	Source      Column   `json:"source"`
	Target      Column   `json:"target"`
	ChangedAttr []string `json:"changed_attributes"`
}

type IndexChange struct {
	Name        string   `json:"name"`
	Source      Index    `json:"source"`
	Target      Index    `json:"target"`
	ChangedAttr []string `json:"changed_attributes"`
}

type ForeignKeyChange struct {
	Name        string     `json:"name"`
	Source      ForeignKey `json:"source"`
	Target      ForeignKey `json:"target"`
	ChangedAttr []string   `json:"changed_attributes"`
}

type ForeignKey struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
	OnUpdate          string
}

type QuerySet struct {
	Schema            string
	Table             string
	Column            string
	Index             string
	ForeignKey        string
	Sequence          string
	SequenceOwnership string
}
