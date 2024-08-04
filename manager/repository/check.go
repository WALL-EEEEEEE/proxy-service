package repository

import (
	"gorm.io/gorm"
)

type CheckItem struct {
	gorm.Model
	Name         string
	Desc         string
	ResourceType string
	ResourceId   string `gorm:"type:varchar(32);index"`
	CheckId      string `gorm:"uniqueIndex;type:varchar(32)"`
}
