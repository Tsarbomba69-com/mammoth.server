package models

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: Password should be encrypted and not stored in plain text
type DBConnection struct {
	gorm.Model
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type Project struct {
	gorm.Model
	Name        string       `json:"name"`
	Description string       `json:"description"`
	SourceID    uint         `json:"source_id"`
	Source      DBConnection `json:"source" gorm:"foreignKey:SourceID;constraint:OnDelete:CASCADE;"`
	TargetID    uint         `json:"target_id"`
	Target      DBConnection `json:"target" gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE;"`
}

// Connect establishes a connection to the database
func (dbc *DBConnection) Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		dbc.Host,
		dbc.User,
		dbc.Password,
		dbc.DBName,
		dbc.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

// ConnectForProject establishes connections to both source and target databases
func (p *Project) ConnectForProject() (*gorm.DB, *gorm.DB, error) {
	// Connect to source database
	sourceDB, err := p.Source.Connect()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to source database: %v", err)
	}

	// Connect to target database
	targetDB, err := p.Target.Connect()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to target database: %v", err)
	}

	return sourceDB, targetDB, nil
}

func (p *Project) GetDialect(db *gorm.DB) string {
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

	return dialect
}
