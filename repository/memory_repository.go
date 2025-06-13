package repository

import (
	"context"
	"familytree/interfaces"
	"familytree/models"
	"fmt"
	"sync"
	"time"
)

// memoryRepository 内存存储库实现（用于演示）
type memoryRepository struct {
	individuals []models.Individual
	families    []models.Family
	children    []models.Child
	events      []models.Event
	places      []models.Place
	sources     []models.Source
	citations   []models.Citation
	notes       []models.Note

	nextIndividualID int
	nextFamilyID     int
	nextChildID      int
	nextEventID      int
	nextPlaceID      int
	nextSourceID     int
	nextCitationID   int
	nextNoteID       int

	mu sync.RWMutex

	// 添加缓存
	individualCache     map[int]*models.Individual
	individualNameCache map[string][]*models.Individual
	cacheMutex          sync.RWMutex
}

// NewMemoryRepository 创建内存存储库
func NewMemoryRepository() interfaces.Repository {
	repo := &memoryRepository{
		individuals: make([]models.Individual, 0),
		families:    make([]models.Family, 0),
		children:    make([]models.Child, 0),
		events:      make([]models.Event, 0),
		places:      make([]models.Place, 0),
		sources:     make([]models.Source, 0),
		citations:   make([]models.Citation, 0),
		notes:       make([]models.Note, 0),

		nextIndividualID: 1,
		nextFamilyID:     1,
		nextChildID:      1,
		nextEventID:      1,
		nextPlaceID:      1,
		nextSourceID:     1,
		nextCitationID:   1,
		nextNoteID:       1,

		// 初始化缓存
		individualCache:     make(map[int]*models.Individual),
		individualNameCache: make(map[string][]*models.Individual),
	}

	// 初始化示例数据
	repo.initSampleData()

	return repo
}

// initSampleData 初始化示例数据
func (r *memoryRepository) initSampleData() {
	now := time.Now()

	// 示例地点
	places := []models.Place{
		{PlaceID: 1, PlaceName: "北京市", Latitude: floatPtr(39.9042), Longitude: floatPtr(116.4074), Notes: "首都", CreatedAt: now, UpdatedAt: now},
		{PlaceID: 2, PlaceName: "上海市", Latitude: floatPtr(31.2304), Longitude: floatPtr(121.4737), Notes: "直辖市", CreatedAt: now, UpdatedAt: now},
		{PlaceID: 3, PlaceName: "广州市", Latitude: floatPtr(23.1291), Longitude: floatPtr(113.2644), Notes: "广东省省会", CreatedAt: now, UpdatedAt: now},
	}
	r.places = places
	r.nextPlaceID = 4

	// 示例个人
	birthDate1950 := time.Date(1950, 1, 15, 0, 0, 0, 0, time.UTC)
	birthDate1955 := time.Date(1955, 3, 20, 0, 0, 0, 0, time.UTC)
	birthDate1975 := time.Date(1975, 6, 10, 0, 0, 0, 0, time.UTC)
	birthDate1978 := time.Date(1978, 9, 15, 0, 0, 0, 0, time.UTC)

	individuals := []models.Individual{
		{IndividualID: 1, FullName: "张伟", Gender: models.GenderMale, BirthDate: &birthDate1950, BirthPlaceID: intPtr(1), Occupation: "工程师", Notes: "家族族长", CreatedAt: now, UpdatedAt: now},
		{IndividualID: 2, FullName: "李丽", Gender: models.GenderFemale, BirthDate: &birthDate1955, BirthPlaceID: intPtr(2), Occupation: "教师", Notes: "张伟的妻子", CreatedAt: now, UpdatedAt: now},
		{IndividualID: 3, FullName: "张明", Gender: models.GenderMale, BirthDate: &birthDate1975, BirthPlaceID: intPtr(3), Occupation: "医生", Notes: "张伟和李丽的儿子", FatherID: intPtr(1), MotherID: intPtr(2), CreatedAt: now, UpdatedAt: now},
		{IndividualID: 4, FullName: "王芳", Gender: models.GenderFemale, BirthDate: &birthDate1978, BirthPlaceID: intPtr(1), Occupation: "律师", Notes: "张明的妻子", CreatedAt: now, UpdatedAt: now},
	}
	r.individuals = individuals
	r.nextIndividualID = 5

	// 示例家庭
	marriageDate1974 := time.Date(1974, 5, 1, 0, 0, 0, 0, time.UTC)
	marriageDate2000 := time.Date(2000, 10, 1, 0, 0, 0, 0, time.UTC)

	families := []models.Family{
		{FamilyID: 1, HusbandID: intPtr(1), WifeID: intPtr(2), MarriageDate: &marriageDate1974, MarriagePlaceID: intPtr(1), Notes: "第一代家庭", CreatedAt: now, UpdatedAt: now},
		{FamilyID: 2, HusbandID: intPtr(3), WifeID: intPtr(4), MarriageDate: &marriageDate2000, MarriagePlaceID: intPtr(2), Notes: "第二代家庭", CreatedAt: now, UpdatedAt: now},
	}
	r.families = families
	r.nextFamilyID = 3

	// 示例子女关系
	children := []models.Child{
		{ChildID: 1, FamilyID: 1, IndividualID: 3, RelationshipToParents: "亲生", CreatedAt: now, UpdatedAt: now},
	}
	r.children = children
	r.nextChildID = 2
}

// 个人信息相关方法
func (r *memoryRepository) CreateIndividual(ctx context.Context, individual *models.Individual) (*models.Individual, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	individual.IndividualID = r.nextIndividualID
	individual.CreatedAt = time.Now()
	individual.UpdatedAt = time.Now()
	r.nextIndividualID++

	r.individuals = append(r.individuals, *individual)

	// 更新缓存
	r.cacheMutex.Lock()
	r.individualCache[individual.IndividualID] = individual
	// 清除名称缓存，因为可能影响搜索结果
	r.individualNameCache = make(map[string][]*models.Individual)
	r.cacheMutex.Unlock()

	return individual, nil
}

func (r *memoryRepository) GetIndividualByID(ctx context.Context, id int) (*models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先从缓存中查找
	r.cacheMutex.RLock()
	if individual, exists := r.individualCache[id]; exists {
		r.cacheMutex.RUnlock()
		return individual, nil
	}
	r.cacheMutex.RUnlock()

	// 缓存未命中，从内存中查找
	for _, individual := range r.individuals {
		if individual.IndividualID == id {
			// 更新缓存
			r.cacheMutex.Lock()
			r.individualCache[id] = &individual
			r.cacheMutex.Unlock()
			return &individual, nil
		}
	}
	return nil, fmt.Errorf("个人不存在")
}

func (r *memoryRepository) UpdateIndividual(ctx context.Context, id int, individual *models.Individual) (*models.Individual, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.individuals {
		if existing.IndividualID == id {
			individual.IndividualID = id
			individual.CreatedAt = existing.CreatedAt
			individual.UpdatedAt = time.Now()
			r.individuals[i] = *individual

			// 更新缓存
			r.cacheMutex.Lock()
			r.individualCache[id] = individual
			// 清除名称缓存，因为可能影响搜索结果
			r.individualNameCache = make(map[string][]*models.Individual)
			r.cacheMutex.Unlock()

			return individual, nil
		}
	}
	return nil, fmt.Errorf("个人不存在")
}

func (r *memoryRepository) DeleteIndividual(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, individual := range r.individuals {
		if individual.IndividualID == id {
			r.individuals = append(r.individuals[:i], r.individuals[i+1:]...)

			// 更新缓存
			r.cacheMutex.Lock()
			delete(r.individualCache, id)
			// 清除名称缓存，因为可能影响搜索结果
			r.individualNameCache = make(map[string][]*models.Individual)
			r.cacheMutex.Unlock()

			return nil
		}
	}
	return fmt.Errorf("个人不存在")
}

func (r *memoryRepository) SearchIndividuals(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	if query == "" {
		// 如果查询为空，直接返回分页结果
		start := offset
		if start > len(r.individuals) {
			start = len(r.individuals)
		}
		end := start + limit
		if end > len(r.individuals) {
			end = len(r.individuals)
		}
		return r.individuals[start:end], len(r.individuals), nil
	}

	// 从缓存中查找
	r.cacheMutex.RLock()
	if results, exists := r.individualNameCache[query]; exists {
		r.cacheMutex.RUnlock()
		// 处理分页
		start := offset
		if start > len(results) {
			start = len(results)
		}
		end := start + limit
		if end > len(results) {
			end = len(results)
		}
		individuals := make([]models.Individual, end-start)
		for i, result := range results[start:end] {
			individuals[i] = *result
		}
		return individuals, len(results), nil
	}
	r.cacheMutex.RUnlock()

	// 缓存未命中，执行搜索
	var results []*models.Individual
	for i := range r.individuals {
		if contains(r.individuals[i].FullName, query) || contains(r.individuals[i].Notes, query) {
			results = append(results, &r.individuals[i])
		}
	}

	// 更新缓存
	r.cacheMutex.Lock()
	r.individualNameCache[query] = results
	r.cacheMutex.Unlock()

	// 处理分页
	start := offset
	if start > len(results) {
		start = len(results)
	}
	end := start + limit
	if end > len(results) {
		end = len(results)
	}

	individuals := make([]models.Individual, end-start)
	for i, result := range results[start:end] {
		individuals[i] = *result
	}

	return individuals, len(results), nil
}

func (r *memoryRepository) GetIndividualsByParentID(ctx context.Context, parentID int) ([]models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var children []models.Individual
	for _, individual := range r.individuals {
		if (individual.FatherID != nil && *individual.FatherID == parentID) ||
			(individual.MotherID != nil && *individual.MotherID == parentID) {
			children = append(children, individual)
		}
	}
	return children, nil
}

func (r *memoryRepository) GetIndividualsByIDs(ctx context.Context, ids []int) ([]models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []models.Individual
	for _, id := range ids {
		for _, individual := range r.individuals {
			if individual.IndividualID == id {
				results = append(results, individual)
				break
			}
		}
	}
	return results, nil
}

func (r *memoryRepository) GetParents(ctx context.Context, individualID int) (*models.Individual, *models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var individual *models.Individual
	for _, ind := range r.individuals {
		if ind.IndividualID == individualID {
			individual = &ind
			break
		}
	}

	if individual == nil {
		return nil, nil, fmt.Errorf("个人不存在")
	}

	var father, mother *models.Individual

	if individual.FatherID != nil {
		for _, ind := range r.individuals {
			if ind.IndividualID == *individual.FatherID {
				father = &ind
				break
			}
		}
	}

	if individual.MotherID != nil {
		for _, ind := range r.individuals {
			if ind.IndividualID == *individual.MotherID {
				mother = &ind
				break
			}
		}
	}

	return father, mother, nil
}

func (r *memoryRepository) GetSiblings(ctx context.Context, individualID int) ([]models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var individual *models.Individual
	for _, ind := range r.individuals {
		if ind.IndividualID == individualID {
			individual = &ind
			break
		}
	}

	if individual == nil {
		return nil, fmt.Errorf("个人不存在")
	}

	var siblings []models.Individual
	for _, ind := range r.individuals {
		if ind.IndividualID == individualID {
			continue
		}

		// 检查是否有相同的父母
		if (individual.FatherID != nil && ind.FatherID != nil && *individual.FatherID == *ind.FatherID) ||
			(individual.MotherID != nil && ind.MotherID != nil && *individual.MotherID == *ind.MotherID) {
			siblings = append(siblings, ind)
		}
	}

	return siblings, nil
}

func (r *memoryRepository) GetSpouses(ctx context.Context, individualID int) ([]models.Individual, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var spouses []models.Individual
	for _, family := range r.families {
		if family.HusbandID != nil && *family.HusbandID == individualID && family.WifeID != nil {
			for _, ind := range r.individuals {
				if ind.IndividualID == *family.WifeID {
					spouses = append(spouses, ind)
					break
				}
			}
		} else if family.WifeID != nil && *family.WifeID == individualID && family.HusbandID != nil {
			for _, ind := range r.individuals {
				if ind.IndividualID == *family.HusbandID {
					spouses = append(spouses, ind)
					break
				}
			}
		}
	}

	return spouses, nil
}

// 简化实现：其他方法返回空或错误
func (r *memoryRepository) GetAncestors(ctx context.Context, individualID int, generations int) ([]models.Individual, error) {
	return []models.Individual{}, nil
}

func (r *memoryRepository) GetDescendants(ctx context.Context, individualID int, generations int) ([]models.Individual, error) {
	return []models.Individual{}, nil
}

// BuildFamilyTree 构建家族树
func (r *memoryRepository) BuildFamilyTree(ctx context.Context, rootID int, generations int) (*models.FamilyTreeNode, error) {
	// 使用缓存来存储已查询的节点
	nodeCache := make(map[int]*models.FamilyTreeNode)

	// 获取根节点
	root, err := r.GetIndividualByID(ctx, rootID)
	if err != nil {
		return nil, err
	}

	// 构建树
	return r.buildTreeRecursive(ctx, root, generations, nodeCache)
}

// buildTreeRecursive 递归构建家族树
func (r *memoryRepository) buildTreeRecursive(ctx context.Context, individual *models.Individual, generations int, nodeCache map[int]*models.FamilyTreeNode) (*models.FamilyTreeNode, error) {
	// 检查缓存
	if node, exists := nodeCache[individual.IndividualID]; exists {
		return node, nil
	}

	// 创建新节点
	node := &models.FamilyTreeNode{
		Individual: individual,
	}

	// 如果还有剩余代数，继续构建
	if generations > 0 {
		// 获取配偶
		spouses, err := r.GetSpouses(ctx, individual.IndividualID)
		if err == nil && len(spouses) > 0 {
			node.Spouse = &spouses[0] // 只取第一个配偶
		}

		// 获取父母
		if individual.FatherID != nil {
			father, err := r.GetIndividualByID(ctx, *individual.FatherID)
			if err == nil {
				node.Parents = append(node.Parents, *father)
			}
		}
		if individual.MotherID != nil {
			mother, err := r.GetIndividualByID(ctx, *individual.MotherID)
			if err == nil {
				node.Parents = append(node.Parents, *mother)
			}
		}

		// 获取子女
		children, err := r.GetIndividualsByParentID(ctx, individual.IndividualID)
		if err == nil {
			for _, child := range children {
				childNode, err := r.buildTreeRecursive(ctx, &child, generations-1, nodeCache)
				if err == nil {
					node.Children = append(node.Children, *childNode)
				}
			}
		}
	}

	// 更新缓存
	nodeCache[individual.IndividualID] = node

	return node, nil
}

// 其他实体的简化实现
func (r *memoryRepository) CreateFamily(ctx context.Context, family *models.Family) (*models.Family, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family.FamilyID = r.nextFamilyID
	family.CreatedAt = time.Now()
	family.UpdatedAt = time.Now()
	r.nextFamilyID++

	r.families = append(r.families, *family)
	return family, nil
}

func (r *memoryRepository) GetFamilyByID(ctx context.Context, id int) (*models.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, family := range r.families {
		if family.FamilyID == id {
			return &family, nil
		}
	}
	return nil, fmt.Errorf("家庭不存在")
}

func (r *memoryRepository) UpdateFamily(ctx context.Context, id int, family *models.Family) (*models.Family, error) {
	return family, nil
}

func (r *memoryRepository) DeleteFamily(ctx context.Context, id int) error {
	return nil
}

func (r *memoryRepository) GetFamiliesByIndividualID(ctx context.Context, individualID int) ([]models.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var families []models.Family
	for _, family := range r.families {
		if (family.HusbandID != nil && *family.HusbandID == individualID) ||
			(family.WifeID != nil && *family.WifeID == individualID) {
			families = append(families, family)
		}
	}
	return families, nil
}

func (r *memoryRepository) CreateChild(ctx context.Context, child *models.Child) (*models.Child, error) {
	return child, nil
}

func (r *memoryRepository) GetChildrenByFamilyID(ctx context.Context, familyID int) ([]models.Child, error) {
	return []models.Child{}, nil
}

func (r *memoryRepository) DeleteChild(ctx context.Context, familyID, childID int) error {
	return nil
}

// 所有其他方法都返回空实现
func (r *memoryRepository) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	return event, nil
}
func (r *memoryRepository) GetEventByID(ctx context.Context, id int) (*models.Event, error) {
	return nil, fmt.Errorf("未实现")
}
func (r *memoryRepository) UpdateEvent(ctx context.Context, id int, event *models.Event) (*models.Event, error) {
	return event, nil
}
func (r *memoryRepository) DeleteEvent(ctx context.Context, id int) error { return nil }
func (r *memoryRepository) GetEventsByIndividualID(ctx context.Context, individualID int) ([]models.Event, error) {
	return []models.Event{}, nil
}
func (r *memoryRepository) GetEventsByType(ctx context.Context, eventType string, limit, offset int) ([]models.Event, int, error) {
	return []models.Event{}, 0, nil
}
func (r *memoryRepository) GetEventsByDateRange(ctx context.Context, startDate, endDate *string, limit, offset int) ([]models.Event, int, error) {
	return []models.Event{}, 0, nil
}
func (r *memoryRepository) GetEventsByPlaceID(ctx context.Context, placeID int, limit, offset int) ([]models.Event, int, error) {
	return []models.Event{}, 0, nil
}

func (r *memoryRepository) CreatePlace(ctx context.Context, place *models.Place) (*models.Place, error) {
	return place, nil
}
func (r *memoryRepository) GetPlaceByID(ctx context.Context, id int) (*models.Place, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, place := range r.places {
		if place.PlaceID == id {
			return &place, nil
		}
	}
	return nil, fmt.Errorf("地点不存在")
}
func (r *memoryRepository) UpdatePlace(ctx context.Context, id int, place *models.Place) (*models.Place, error) {
	return place, nil
}
func (r *memoryRepository) DeletePlace(ctx context.Context, id int) error { return nil }
func (r *memoryRepository) SearchPlaces(ctx context.Context, query string, limit, offset int) ([]models.Place, int, error) {
	return r.places, len(r.places), nil
}
func (r *memoryRepository) GetPlacesByCoordinates(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit, offset int) ([]models.Place, int, error) {
	return []models.Place{}, 0, nil
}

func (r *memoryRepository) CreateSource(ctx context.Context, source *models.Source) (*models.Source, error) {
	return source, nil
}
func (r *memoryRepository) GetSourceByID(ctx context.Context, id int) (*models.Source, error) {
	return nil, fmt.Errorf("未实现")
}
func (r *memoryRepository) UpdateSource(ctx context.Context, id int, source *models.Source) (*models.Source, error) {
	return source, nil
}
func (r *memoryRepository) DeleteSource(ctx context.Context, id int) error { return nil }
func (r *memoryRepository) SearchSources(ctx context.Context, query string, limit, offset int) ([]models.Source, int, error) {
	return []models.Source{}, 0, nil
}
func (r *memoryRepository) GetSourcesByAuthor(ctx context.Context, author string, limit, offset int) ([]models.Source, int, error) {
	return []models.Source{}, 0, nil
}
func (r *memoryRepository) GetSourcesByYear(ctx context.Context, year int, limit, offset int) ([]models.Source, int, error) {
	return []models.Source{}, 0, nil
}

func (r *memoryRepository) CreateCitation(ctx context.Context, citation *models.Citation) (*models.Citation, error) {
	return citation, nil
}
func (r *memoryRepository) GetCitationByID(ctx context.Context, id int) (*models.Citation, error) {
	return nil, fmt.Errorf("未实现")
}
func (r *memoryRepository) UpdateCitation(ctx context.Context, id int, citation *models.Citation) (*models.Citation, error) {
	return citation, nil
}
func (r *memoryRepository) DeleteCitation(ctx context.Context, id int) error { return nil }
func (r *memoryRepository) GetCitationsByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Citation, error) {
	return []models.Citation{}, nil
}
func (r *memoryRepository) GetCitationsBySourceID(ctx context.Context, sourceID int, limit, offset int) ([]models.Citation, int, error) {
	return []models.Citation{}, 0, nil
}

func (r *memoryRepository) CreateNote(ctx context.Context, note *models.Note) (*models.Note, error) {
	return note, nil
}
func (r *memoryRepository) GetNoteByID(ctx context.Context, id int) (*models.Note, error) {
	return nil, fmt.Errorf("未实现")
}
func (r *memoryRepository) UpdateNote(ctx context.Context, id int, note *models.Note) (*models.Note, error) {
	return note, nil
}
func (r *memoryRepository) DeleteNote(ctx context.Context, id int) error { return nil }
func (r *memoryRepository) GetNotesByEntity(ctx context.Context, entityType models.EntityType, entityID int) ([]models.Note, error) {
	return []models.Note{}, nil
}
func (r *memoryRepository) SearchNotes(ctx context.Context, query string, limit, offset int) ([]models.Note, int, error) {
	return []models.Note{}, 0, nil
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
