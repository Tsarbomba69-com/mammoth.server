package ddl

import (
	"github.com/Tsarbomba69-com/mammoth.server/models"
	"gorm.io/gorm"
)

type DDL interface {
	CreateTableSQL(tableDiff models.TableDiff) string
	AlterTableSQL(tableDiff models.TableDiff) string
	RevertAlterTableSQL(tableDiff models.TableDiff) string
	CreateIndexSQL(tableName string, idx models.IndexInfo) string
	DropIndexSQL(tableName string, idx models.IndexInfo) string
	AddForeignKeySQL(table string, fk models.ForeignKeyInfo) string
	DropForeignKeySQL(table, constraint string) string
	DropTableSQL(tableName string) string
	DumpDatabaseSQL(connection models.DBConnection, db *gorm.DB) (string, error)
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
