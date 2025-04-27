package schemas

import (
	"time"

	"github.com/Tsarbomba69-com/mammoth.server/services"
	"github.com/Tsarbomba69-com/mammoth.server/types"
)

type DBConnectionRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	DBName   string `json:"dbname" binding:"required"`
}

type ProjectRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	Source      DBConnectionRequest `json:"source" binding:"required"`
	Target      DBConnectionRequest `json:"target" binding:"required"`
}

type DBConnectionResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	User      string    `json:"user"`
	DBName    string    `json:"dbname"`
}

type ProjectResponse struct {
	ID          uint                 `json:"id"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Source      DBConnectionResponse `json:"source"`
	Target      DBConnectionResponse `json:"target"`
}

type SchemaComparisonResponse struct {
	Differences     types.SchemaDiff         `json:"differences"`
	MigrationScript services.MigrationScript `json:"migration_script"`
}
