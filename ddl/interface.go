package ddl

import (
	"github.com/Tsarbomba69-com/mammoth.server/types"
)

type DDL interface {
	CreateTableSQL(tableDiff types.TableDiff) string
	AlterTableSQL(tableDiff types.TableDiff) string
	RevertAlterTableSQL(tableDiff types.TableDiff) string
	CreateIndexSQL(tableName string, idx types.IndexInfo) string
	DropIndexSQL(tableName string, idx types.IndexInfo) string
	AddForeignKeySQL(table string, fk types.ForeignKeyInfo) string
	DropForeignKeySQL(table, constraint string) string
	DropTableSQL(tableName string) string
}

func NewDDL(dialect string) DDL {
	switch dialect {
	// case "mysql":
	//     return MySQLDDL{}
	case "postgres":
		return PostgreSQLDDL{}
	// case "sqlite":
	//     return SQLiteDDL{}
	// case "sqlserver":
	//     return SQLServerDDL{}
	default:
		panic("unsupported dialect")
	}
}
