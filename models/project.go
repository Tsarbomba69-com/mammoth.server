package models

import (
	"gorm.io/gorm"
)

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
