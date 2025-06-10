package models

import (
	"database/sql/driver"
	"time"
)

// Gender 性别枚举
type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderOther   Gender = "other"
	GenderUnknown Gender = "unknown"
)

// Scan 实现 sql.Scanner 接口
func (g *Gender) Scan(value interface{}) error {
	if value == nil {
		*g = GenderUnknown
		return nil
	}
	switch s := value.(type) {
	case string:
		*g = Gender(s)
	case []byte:
		*g = Gender(s)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (g Gender) Value() (driver.Value, error) {
	return string(g), nil
}

// EntityType 实体类型枚举
type EntityType string

const (
	EntityTypeIndividual EntityType = "Individual"
	EntityTypeFamily     EntityType = "Family"
	EntityTypeEvent      EntityType = "Event"
	EntityTypeSource     EntityType = "Source"
	EntityTypePlace      EntityType = "Place"
)

// Individual 个人信息结构体
type Individual struct {
	IndividualID   int       `json:"individual_id" db:"individual_id"`
	FullName       string    `json:"full_name" db:"full_name"`
	Gender         Gender    `json:"gender" db:"gender"`
	BirthDate      *time.Time `json:"birth_date" db:"birth_date"`
	BirthPlace     string    `json:"birth_place" db:"birth_place"`           // 出生地点（文本）
	BirthPlaceID   *int      `json:"birth_place_id" db:"birth_place_id"`
	DeathDate      *time.Time `json:"death_date" db:"death_date"`
	DeathPlace     string    `json:"death_place" db:"death_place"`           // 埋葬地点（文本）
	BurialPlace    string    `json:"burial_place" db:"burial_place"`         // 埋葬地点（别名，兼容性）
	DeathPlaceID   *int      `json:"death_place_id" db:"death_place_id"`
	Occupation     string    `json:"occupation" db:"occupation"`
	Notes          string    `json:"notes" db:"notes"`
	PhotoURL       string    `json:"photo_url" db:"photo_url"`
	FatherID       *int      `json:"father_id" db:"father_id"`
	MotherID       *int      `json:"mother_id" db:"mother_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	
	// 关联字段（非数据库字段）
	BirthPlaceObj  *Place      `json:"birth_place_obj,omitempty" db:"-"`
	DeathPlaceObj  *Place      `json:"death_place_obj,omitempty" db:"-"`
	Father         *Individual `json:"father,omitempty" db:"-"`
	Mother         *Individual `json:"mother,omitempty" db:"-"`
	Children       []Individual `json:"children,omitempty" db:"-"`
}

// Family 家庭关系结构体
type Family struct {
	FamilyID         int       `json:"family_id" db:"family_id"`
	HusbandID        *int      `json:"husband_id" db:"husband_id"`
	WifeID           *int      `json:"wife_id" db:"wife_id"`
	MarriageDate     *time.Time `json:"marriage_date" db:"marriage_date"`
	MarriagePlaceID  *int      `json:"marriage_place_id" db:"marriage_place_id"`
	DivorceDate      *time.Time `json:"divorce_date" db:"divorce_date"`
	Notes            string    `json:"notes" db:"notes"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	
	// 关联字段（非数据库字段）
	Husband          *Individual `json:"husband,omitempty" db:"-"`
	Wife             *Individual `json:"wife,omitempty" db:"-"`
	MarriagePlace    *Place      `json:"marriage_place,omitempty" db:"-"`
	Children         []Child     `json:"children,omitempty" db:"-"`
}

// Child 子女关系结构体
type Child struct {
	ChildID                int       `json:"child_id" db:"child_id"`
	FamilyID               int       `json:"family_id" db:"family_id"`
	IndividualID           int       `json:"individual_id" db:"individual_id"`
	RelationshipToParents  string    `json:"relationship_to_parents" db:"relationship_to_parents"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
	
	// 关联字段（非数据库字段）
	Individual             *Individual `json:"individual,omitempty" db:"-"`
	Family                 *Family     `json:"family,omitempty" db:"-"`
}

// Event 事件结构体
type Event struct {
	EventID        int       `json:"event_id" db:"event_id"`
	IndividualID   int       `json:"individual_id" db:"individual_id"`
	EventType      string    `json:"event_type" db:"event_type"`
	EventDate      *time.Time `json:"event_date" db:"event_date"`
	EventPlaceID   *int      `json:"event_place_id" db:"event_place_id"`
	Description    string    `json:"description" db:"description"`
	Notes          string    `json:"notes" db:"notes"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	
	// 关联字段（非数据库字段）
	Individual     *Individual `json:"individual,omitempty" db:"-"`
	EventPlace     *Place      `json:"event_place,omitempty" db:"-"`
}

// Place 地点结构体
type Place struct {
	PlaceID     int       `json:"place_id" db:"place_id"`
	PlaceName   string    `json:"place_name" db:"place_name"`
	Latitude    *float64  `json:"latitude" db:"latitude"`
	Longitude   *float64  `json:"longitude" db:"longitude"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Source 信息来源结构体
type Source struct {
	SourceID        int       `json:"source_id" db:"source_id"`
	Title           string    `json:"title" db:"title"`
	Author          string    `json:"author" db:"author"`
	PublicationYear *int      `json:"publication_year" db:"publication_year"`
	Publisher       string    `json:"publisher" db:"publisher"`
	Location        string    `json:"location" db:"location"`
	Notes           string    `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Citation 引用结构体
type Citation struct {
	CitationID  int        `json:"citation_id" db:"citation_id"`
	SourceID    int        `json:"source_id" db:"source_id"`
	EntityType  EntityType `json:"entity_type" db:"entity_type"`
	EntityID    int        `json:"entity_id" db:"entity_id"`
	PageNumber  string     `json:"page_number" db:"page_number"`
	Notes       string     `json:"notes" db:"notes"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	
	// 关联字段（非数据库字段）
	Source      *Source    `json:"source,omitempty" db:"-"`
}

// Note 通用备注结构体
type Note struct {
	NoteID     int        `json:"note_id" db:"note_id"`
	EntityType EntityType `json:"entity_type" db:"entity_type"`
	EntityID   int        `json:"entity_id" db:"entity_id"`
	NoteText   string     `json:"note_text" db:"note_text"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateIndividualRequest 创建个人信息请求
type CreateIndividualRequest struct {
	FullName     string     `json:"full_name" binding:"required"`
	Gender       Gender     `json:"gender"`
	BirthDate    *time.Time `json:"birth_date"`
	BirthPlace   string     `json:"birth_place"`    // 出生地点
	BirthPlaceID *int       `json:"birth_place_id"`
	DeathDate    *time.Time `json:"death_date"`
	DeathPlace   string     `json:"death_place"`    // 埋葬地点
	BurialPlace  string     `json:"burial_place"`   // 埋葬地点（别名）
	DeathPlaceID *int       `json:"death_place_id"`
	Occupation   string     `json:"occupation"`
	Notes        string     `json:"notes"`
	PhotoURL     string     `json:"photo_url"`
	FatherID     *int       `json:"father_id"`
	MotherID     *int       `json:"mother_id"`
}

// UpdateIndividualRequest 更新个人信息请求
type UpdateIndividualRequest struct {
	FullName     *string    `json:"full_name"`
	Gender       *Gender    `json:"gender"`
	BirthDate    *time.Time `json:"birth_date"`
	BirthPlace   *string    `json:"birth_place"`    // 出生地点
	BirthPlaceID *int       `json:"birth_place_id"`
	DeathDate    *time.Time `json:"death_date"`
	DeathPlace   *string    `json:"death_place"`    // 埋葬地点
	BurialPlace  *string    `json:"burial_place"`   // 埋葬地点（别名）
	DeathPlaceID *int       `json:"death_place_id"`
	Occupation   *string    `json:"occupation"`
	Notes        *string    `json:"notes"`
	PhotoURL     *string    `json:"photo_url"`
	FatherID     *int       `json:"father_id"`
	MotherID     *int       `json:"mother_id"`
}

// CreateFamilyRequest 创建家庭关系请求
type CreateFamilyRequest struct {
	HusbandID        *int       `json:"husband_id"`
	WifeID           *int       `json:"wife_id"`
	MarriageDate     *time.Time `json:"marriage_date"`
	MarriagePlaceID  *int       `json:"marriage_place_id"`
	DivorceDate      *time.Time `json:"divorce_date"`
	Notes            string     `json:"notes"`
}

// FamilyTreeNode 家族树节点
type FamilyTreeNode struct {
	Individual *Individual      `json:"individual"`
	Spouse     *Individual      `json:"spouse,omitempty"`
	Children   []FamilyTreeNode `json:"children,omitempty"`
	Parents    []Individual     `json:"parents,omitempty"`
}

// PaginationResponse 分页响应结构体
type PaginationResponse struct {
	Data   interface{} `json:"data"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
} 