package interfaces

import (
	"context"
	"familytree/models"
)

// IndividualService 个人信息服务接口
type IndividualService interface {
	// 创建个人信息
	Create(ctx context.Context, req *models.CreateIndividualRequest) (*models.Individual, error)

	// 根据ID获取个人信息
	GetByID(ctx context.Context, id int) (*models.Individual, error)

	// 更新个人信息
	Update(ctx context.Context, id int, req *models.UpdateIndividualRequest) (*models.Individual, error)

	// 删除个人信息
	Delete(ctx context.Context, id int) error

	// 搜索个人信息
	Search(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error)

	// 获取个人的所有子女
	GetChildren(ctx context.Context, id int) ([]models.Individual, error)

	// 获取个人的父母
	GetParents(ctx context.Context, id int) (father, mother *models.Individual, err error)

	// 获取个人的兄弟姐妹
	GetSiblings(ctx context.Context, id int) ([]models.Individual, error)

	// 获取个人的配偶
	GetSpouses(ctx context.Context, id int) ([]models.Individual, error)

	// 获取个人的所有祖先
	GetAncestors(ctx context.Context, id int, generations int) ([]models.Individual, error)

	// 获取个人的所有后代
	GetDescendants(ctx context.Context, id int, generations int) ([]models.Individual, error)

	// 获取家族树
	GetFamilyTree(ctx context.Context, rootID int, generations int) (*models.FamilyTreeNode, error)

	// 向上添加父母
	AddParent(ctx context.Context, childID int, req *models.AddParentRequest) (*models.Individual, error)
}

// FamilyService 家庭关系服务接口
type FamilyService interface {
	// 创建家庭关系
	CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error)

	// 根据ID获取家庭关系
	GetByID(ctx context.Context, id int) (*models.Family, error)

	// 更新家庭关系
	Update(ctx context.Context, id int, req *models.CreateFamilyRequest) (*models.Family, error)

	// 删除家庭关系
	Delete(ctx context.Context, id int) error

	// 根据夫妻ID获取家庭关系
	GetBySpouses(ctx context.Context, husbandID, wifeID int) (*models.Family, error)

	// 获取某人参与的所有家庭关系
	GetByIndividualID(ctx context.Context, individualID int) ([]models.Family, error)

	// 添加配偶关系
	AddSpouse(ctx context.Context, individualID, spouseID int) (*models.Family, error)

	// 为家庭添加子女
	AddChild(ctx context.Context, familyID, childID int, relationship string) error

	// 从家庭移除子女
	RemoveChild(ctx context.Context, familyID, childID int) error

	// 获取家庭的所有子女
	GetChildren(ctx context.Context, familyID int) ([]models.Child, error)
}

// EventService 事件服务接口
type EventService interface {
	// 创建事件
	Create(ctx context.Context, event *models.Event) (*models.Event, error)

	// 根据ID获取事件
	GetByID(ctx context.Context, id int) (*models.Event, error)

	// 更新事件
	Update(ctx context.Context, id int, event *models.Event) (*models.Event, error)

	// 删除事件
	Delete(ctx context.Context, id int) error

	// 获取个人的所有事件
	GetByIndividualID(ctx context.Context, individualID int) ([]models.Event, error)

	// 根据事件类型获取事件
	GetByType(ctx context.Context, eventType string, limit, offset int) ([]models.Event, int, error)

	// 根据日期范围获取事件
	GetByDateRange(ctx context.Context, startDate, endDate *string, limit, offset int) ([]models.Event, int, error)

	// 根据地点获取事件
	GetByPlace(ctx context.Context, placeID int, limit, offset int) ([]models.Event, int, error)
}

// PlaceService 地点服务接口
type PlaceService interface {
	// 创建地点
	Create(ctx context.Context, place *models.Place) (*models.Place, error)

	// 根据ID获取地点
	GetByID(ctx context.Context, id int) (*models.Place, error)

	// 更新地点
	Update(ctx context.Context, id int, place *models.Place) (*models.Place, error)

	// 删除地点
	Delete(ctx context.Context, id int) error

	// 搜索地点
	Search(ctx context.Context, query string, limit, offset int) ([]models.Place, int, error)

	// 根据坐标范围获取地点
	GetByCoordinates(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit, offset int) ([]models.Place, int, error)
}

// SourceService 信息来源服务接口
type SourceService interface {
	// 创建信息来源
	Create(ctx context.Context, source *models.Source) (*models.Source, error)

	// 根据ID获取信息来源
	GetByID(ctx context.Context, id int) (*models.Source, error)

	// 更新信息来源
	Update(ctx context.Context, id int, source *models.Source) (*models.Source, error)

	// 删除信息来源
	Delete(ctx context.Context, id int) error

	// 搜索信息来源
	Search(ctx context.Context, query string, limit, offset int) ([]models.Source, int, error)

	// 根据作者获取信息来源
	GetByAuthor(ctx context.Context, author string, limit, offset int) ([]models.Source, int, error)

	// 根据出版年份获取信息来源
	GetByYear(ctx context.Context, year int, limit, offset int) ([]models.Source, int, error)
}

// CitationService 引用服务接口
type CitationService interface {
	// 创建引用
	Create(ctx context.Context, citation *models.Citation) (*models.Citation, error)

	// 根据ID获取引用
	GetByID(ctx context.Context, id int) (*models.Citation, error)

	// 更新引用
	Update(ctx context.Context, id int, citation *models.Citation) (*models.Citation, error)

	// 删除引用
	Delete(ctx context.Context, id int) error

	// 根据实体获取引用
	GetByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Citation, error)

	// 根据来源获取引用
	GetBySource(ctx context.Context, sourceID int, limit, offset int) ([]models.Citation, int, error)
}

// NoteService 备注服务接口
type NoteService interface {
	// 创建备注
	Create(ctx context.Context, note *models.Note) (*models.Note, error)

	// 根据ID获取备注
	GetByID(ctx context.Context, id int) (*models.Note, error)

	// 更新备注
	Update(ctx context.Context, id int, note *models.Note) (*models.Note, error)

	// 删除备注
	Delete(ctx context.Context, id int) error

	// 根据实体获取备注
	GetByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Note, error)

	// 搜索备注
	Search(ctx context.Context, query string, limit, offset int) ([]models.Note, int, error)
}

// Repository 数据访问层接口
type Repository interface {
	IndividualRepository
	FamilyRepository
	EventRepository
	PlaceRepository
	SourceRepository
	CitationRepository
	NoteRepository
}

// IndividualRepository 个人信息数据访问接口
type IndividualRepository interface {
	CreateIndividual(ctx context.Context, individual *models.Individual) (*models.Individual, error)
	GetIndividualByID(ctx context.Context, id int) (*models.Individual, error)
	UpdateIndividual(ctx context.Context, id int, individual *models.Individual) (*models.Individual, error)
	DeleteIndividual(ctx context.Context, id int) error
	SearchIndividuals(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error)
	GetIndividualsByParentID(ctx context.Context, parentID int) ([]models.Individual, error)
	GetIndividualsByIDs(ctx context.Context, ids []int) ([]models.Individual, error)
	GetSpouses(ctx context.Context, individualID int) ([]models.Individual, error)
}

// FamilyRepository 家庭关系数据访问接口
type FamilyRepository interface {
	CreateFamily(ctx context.Context, family *models.Family) (*models.Family, error)
	GetFamilyByID(ctx context.Context, id int) (*models.Family, error)
	UpdateFamily(ctx context.Context, id int, family *models.Family) (*models.Family, error)
	DeleteFamily(ctx context.Context, id int) error
	GetFamiliesByIndividualID(ctx context.Context, individualID int) ([]models.Family, error)
	CreateChild(ctx context.Context, child *models.Child) (*models.Child, error)
	DeleteChild(ctx context.Context, familyID, individualID int) error
	GetChildrenByFamilyID(ctx context.Context, familyID int) ([]models.Child, error)
}

// EventRepository 事件数据访问接口
type EventRepository interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	GetEventByID(ctx context.Context, id int) (*models.Event, error)
	UpdateEvent(ctx context.Context, id int, event *models.Event) (*models.Event, error)
	DeleteEvent(ctx context.Context, id int) error
	GetEventsByIndividualID(ctx context.Context, individualID int) ([]models.Event, error)
	GetEventsByType(ctx context.Context, eventType string, limit, offset int) ([]models.Event, int, error)
	GetEventsByDateRange(ctx context.Context, startDate, endDate *string, limit, offset int) ([]models.Event, int, error)
	GetEventsByPlaceID(ctx context.Context, placeID int, limit, offset int) ([]models.Event, int, error)
}

// PlaceRepository 地点数据访问接口
type PlaceRepository interface {
	CreatePlace(ctx context.Context, place *models.Place) (*models.Place, error)
	GetPlaceByID(ctx context.Context, id int) (*models.Place, error)
	UpdatePlace(ctx context.Context, id int, place *models.Place) (*models.Place, error)
	DeletePlace(ctx context.Context, id int) error
	SearchPlaces(ctx context.Context, query string, limit, offset int) ([]models.Place, int, error)
	GetPlacesByCoordinates(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit, offset int) ([]models.Place, int, error)
}

// SourceRepository 信息来源数据访问接口
type SourceRepository interface {
	CreateSource(ctx context.Context, source *models.Source) (*models.Source, error)
	GetSourceByID(ctx context.Context, id int) (*models.Source, error)
	UpdateSource(ctx context.Context, id int, source *models.Source) (*models.Source, error)
	DeleteSource(ctx context.Context, id int) error
	SearchSources(ctx context.Context, query string, limit, offset int) ([]models.Source, int, error)
	GetSourcesByAuthor(ctx context.Context, author string, limit, offset int) ([]models.Source, int, error)
	GetSourcesByYear(ctx context.Context, year int, limit, offset int) ([]models.Source, int, error)
}

// CitationRepository 引用数据访问接口
type CitationRepository interface {
	CreateCitation(ctx context.Context, citation *models.Citation) (*models.Citation, error)
	GetCitationByID(ctx context.Context, id int) (*models.Citation, error)
	UpdateCitation(ctx context.Context, id int, citation *models.Citation) (*models.Citation, error)
	DeleteCitation(ctx context.Context, id int) error
	GetCitationsByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Citation, error)
	GetCitationsBySourceID(ctx context.Context, sourceID int, limit, offset int) ([]models.Citation, int, error)
}

// NoteRepository 备注数据访问接口
type NoteRepository interface {
	CreateNote(ctx context.Context, note *models.Note) (*models.Note, error)
	GetNoteByID(ctx context.Context, id int) (*models.Note, error)
	UpdateNote(ctx context.Context, id int, note *models.Note) (*models.Note, error)
	DeleteNote(ctx context.Context, id int) error
	GetNotesByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Note, error)
	SearchNotes(ctx context.Context, query string, limit, offset int) ([]models.Note, int, error)
}
