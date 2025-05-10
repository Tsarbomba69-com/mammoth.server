package ddl

import (
	"github.com/Tsarbomba69-com/mammoth.server/models"
	"gorm.io/gorm"
)

type DDL interface {
	CreateSchemaSQL(schemaName string) string
	CreateTableSQL(tableDiff models.TableDiff) string
	AlterTableSQL(tableDiff models.TableDiff) string
	RevertAlterTableSQL(tableDiff models.TableDiff) string
	CreateIndexSQL(schemaName, tableName string, idx models.Index) string
	DropIndexSQL(schemaName, tableName string, idx models.Index) string
	AddForeignKeySQL(schemaName, tableName string, fk models.ForeignKey) string
	DropForeignKeySQL(schemaName, tableName, constraint string) string
	DropTableSQL(schemaName, tableName string) string
	DumpDatabaseSQL(connection models.DBConnection, db *gorm.DB) (string, error)
	DropSchemaSQL(schema string) string
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
