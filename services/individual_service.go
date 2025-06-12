package services

import (
	"context"
	"fmt"
	"time"

	"familytree/interfaces"
	"familytree/models"
)

// IndividualService 个人信息服务
type IndividualService struct {
	repo       interfaces.IndividualRepository
	familyRepo interfaces.FamilyRepository
}

// NewIndividualService 创建个人信息服务
func NewIndividualService(repo interfaces.IndividualRepository, familyRepo interfaces.FamilyRepository) interfaces.IndividualService {
	return &IndividualService{
		repo:       repo,
		familyRepo: familyRepo,
	}
}

// Create 创建个人信息
func (s *IndividualService) Create(ctx context.Context, req *models.CreateIndividualRequest) (*models.Individual, error) {
	// 验证必填字段
	if req.FullName == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}

	// 验证父母关系
	if req.FatherID != nil && req.MotherID != nil && *req.FatherID == *req.MotherID {
		return nil, fmt.Errorf("父亲和母亲不能是同一个人")
	}

	// 如果指定了父亲，验证父亲存在且为男性
	if req.FatherID != nil {
		father, err := s.repo.GetIndividualByID(ctx, *req.FatherID)
		if err != nil {
			return nil, fmt.Errorf("父亲不存在")
		}
		if father.Gender != models.GenderMale {
			return nil, fmt.Errorf("指定的父亲必须是男性")
		}

		// 如果指定了母亲，验证母亲存在且为女性
		if req.MotherID != nil {
			mother, err := s.repo.GetIndividualByID(ctx, *req.MotherID)
			if err != nil {
				return nil, fmt.Errorf("母亲不存在")
			}
			if mother.Gender != models.GenderFemale {
				return nil, fmt.Errorf("指定的母亲必须是女性")
			}

			// 验证父母是否已婚，如果没有则自动创建婚姻关系
			families, err := s.familyRepo.GetFamiliesByIndividualID(ctx, *req.FatherID)
			if err != nil {
				return nil, fmt.Errorf("验证父母婚姻关系失败: %v", err)
			}

			married := false
			for _, family := range families {
				if family.HusbandID != nil && *family.HusbandID == *req.FatherID &&
					family.WifeID != nil && *family.WifeID == *req.MotherID {
					married = true
					break
				}
			}

			// 如果父母未建立婚姻关系，自动创建
			if !married {
				// 获取父亲的下一个婚姻顺序
				marriageOrder := 1
				for _, family := range families {
					if family.HusbandID != nil && *family.HusbandID == *req.FatherID {
						if family.MarriageOrder >= marriageOrder {
							marriageOrder = family.MarriageOrder + 1
						}
					}
				}

				// 创建婚姻关系
				newFamily := &models.Family{
					HusbandID:     req.FatherID,
					WifeID:        req.MotherID,
					MarriageOrder: marriageOrder,
					Notes:         "系统自动创建的婚姻关系",
				}

				_, err := s.familyRepo.CreateFamily(ctx, newFamily)
				if err != nil {
					return nil, fmt.Errorf("创建父母婚姻关系失败: %v", err)
				}
			}
		}
	} else if req.MotherID != nil {
		// 如果只指定了母亲，验证母亲存在且为女性
		mother, err := s.repo.GetIndividualByID(ctx, *req.MotherID)
		if err != nil {
			return nil, fmt.Errorf("母亲不存在")
		}
		if mother.Gender != models.GenderFemale {
			return nil, fmt.Errorf("指定的母亲必须是女性")
		}
	}

	// 创建个人信息
	individual := &models.Individual{
		FullName:     req.FullName,
		Gender:       req.Gender,
		BirthDate:    req.BirthDate,
		BirthPlace:   req.BirthPlace,
		BirthPlaceID: req.BirthPlaceID,
		DeathDate:    req.DeathDate,
		DeathPlace:   req.DeathPlace,
		BurialPlace:  req.BurialPlace,
		DeathPlaceID: req.DeathPlaceID,
		Occupation:   req.Occupation,
		Notes:        req.Notes,
		PhotoURL:     req.PhotoURL,
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdIndividual, err := s.repo.CreateIndividual(ctx, individual)
	if err != nil {
		return nil, err
	}

	// 如果有父母，创建子女关系记录
	if req.FatherID != nil && req.MotherID != nil {
		// 查找父母的家庭关系
		families, err := s.familyRepo.GetFamiliesByIndividualID(ctx, *req.FatherID)
		if err == nil {
			for _, family := range families {
				if family.HusbandID != nil && *family.HusbandID == *req.FatherID &&
					family.WifeID != nil && *family.WifeID == *req.MotherID {
					// 创建子女关系记录
					child := &models.Child{
						FamilyID:               family.FamilyID,
						IndividualID:           createdIndividual.IndividualID,
						RelationshipToParents:  "生子",
					}
					if createdIndividual.Gender == models.GenderFemale {
						child.RelationshipToParents = "生女"
					}
					
					s.familyRepo.CreateChild(ctx, child)
					break
				}
			}
		}
	}

	return createdIndividual, nil
}

// GetByID 根据ID获取个人信息
func (s *IndividualService) GetByID(ctx context.Context, id int) (*models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}
	return s.repo.GetIndividualByID(ctx, id)
}

// Update 更新个人信息
func (s *IndividualService) Update(ctx context.Context, id int, req *models.UpdateIndividualRequest) (*models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}

	// 验证输入
	if req.FullName != nil && *req.FullName == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}

	// 验证不能将自己设为父母
	if req.FatherID != nil && *req.FatherID == id {
		return nil, fmt.Errorf("不能将自己设为父亲")
	}
	if req.MotherID != nil && *req.MotherID == id {
		return nil, fmt.Errorf("不能将自己设为母亲")
	}

	// 验证父母关系
	if req.FatherID != nil && req.MotherID != nil && *req.FatherID == *req.MotherID {
		return nil, fmt.Errorf("父亲和母亲不能是同一个人")
	}

	// 验证循环关系 - 防止A的父亲是B，B的父亲是A这种情况
	if req.FatherID != nil {
		if err := s.validateNoCircularRelationship(ctx, id, *req.FatherID, "父亲"); err != nil {
			return nil, err
		}
	}
	if req.MotherID != nil {
		if err := s.validateNoCircularRelationship(ctx, id, *req.MotherID, "母亲"); err != nil {
			return nil, err
		}
	}

	// 验证父亲性别
	if req.FatherID != nil {
		father, err := s.repo.GetIndividualByID(ctx, *req.FatherID)
		if err != nil {
			return nil, fmt.Errorf("父亲信息不存在")
		}
		if father.Gender != models.GenderMale {
			return nil, fmt.Errorf("父亲必须是男性")
		}
	}

	// 验证母亲性别
	if req.MotherID != nil {
		mother, err := s.repo.GetIndividualByID(ctx, *req.MotherID)
		if err != nil {
			return nil, fmt.Errorf("母亲信息不存在")
		}
		if mother.Gender != models.GenderFemale {
			return nil, fmt.Errorf("母亲必须是女性")
		}
	}

	// 获取当前个人信息用于合并更新
	current, err := s.repo.GetIndividualByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查性别是否变更
	var newGender models.Gender
	if req.Gender != nil {
		newGender = *req.Gender
	} else {
		newGender = current.Gender
	}

	// 如果性别变更了，需要调整相关的父母关系
	if current.Gender != newGender {
		// 获取以此人为父亲/母亲的所有子女
		children, err := s.repo.GetIndividualsByParentID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("检查子女关系失败: %v", err)
		}

		// 根据性别变更调整子女的父母关系
		for _, child := range children {
			childUpdate := &models.Individual{
				FullName:     child.FullName,
				Gender:       child.Gender,
				BirthDate:    child.BirthDate,
				BirthPlace:   child.BirthPlace,
				BirthPlaceID: child.BirthPlaceID,
				DeathDate:    child.DeathDate,
				DeathPlace:   child.DeathPlace,
				BurialPlace:  child.BurialPlace,
				DeathPlaceID: child.DeathPlaceID,
				Occupation:   child.Occupation,
				Notes:        child.Notes,
				PhotoURL:     child.PhotoURL,
				FatherID:     child.FatherID,
				MotherID:     child.MotherID,
			}

			// 如果从男性变为女性，将此人从父亲转为母亲
			if current.Gender == models.GenderMale && newGender != models.GenderMale {
				if child.FatherID != nil && *child.FatherID == id {
					childUpdate.FatherID = nil
					childUpdate.MotherID = &id
				}
			}

			// 如果从女性变为男性，将此人从母亲转为父亲
			if current.Gender == models.GenderFemale && newGender != models.GenderFemale {
				if child.MotherID != nil && *child.MotherID == id {
					childUpdate.MotherID = nil
					childUpdate.FatherID = &id
				}
			}

			// 更新子女信息
			_, err = s.repo.UpdateIndividual(ctx, child.IndividualID, childUpdate)
			if err != nil {
				return nil, fmt.Errorf("更新子女关系失败: %v", err)
			}
		}
	}

	// 合并更新字段
	if req.FatherID == nil {
		req.FatherID = current.FatherID
	}
	if req.MotherID == nil {
		req.MotherID = current.MotherID
	}
	if req.BirthDate == nil {
		req.BirthDate = current.BirthDate
	}
	if req.BirthPlace == nil {
		birthPlace := current.BirthPlace
		req.BirthPlace = &birthPlace
	}
	if req.BirthPlaceID == nil {
		req.BirthPlaceID = current.BirthPlaceID
	}
	if req.DeathDate == nil {
		req.DeathDate = current.DeathDate
	}
	if req.DeathPlace == nil {
		deathPlace := current.DeathPlace
		req.DeathPlace = &deathPlace
	}
	if req.BurialPlace == nil {
		burialPlace := current.BurialPlace
		req.BurialPlace = &burialPlace
	}
	if req.DeathPlaceID == nil {
		req.DeathPlaceID = current.DeathPlaceID
	}
	if req.Occupation == nil {
		occupation := current.Occupation
		req.Occupation = &occupation
	}
	if req.Notes == nil {
		notes := current.Notes
		req.Notes = &notes
	}
	if req.PhotoURL == nil {
		photoURL := current.PhotoURL
		req.PhotoURL = photoURL
	}

	// 更新个人信息（使用指针字段的值或保持原值）
	individual := &models.Individual{
		FullName:     getStringValue(req.FullName, current.FullName),
		Gender:       getGenderValue(req.Gender, current.Gender),
		BirthDate:    req.BirthDate,
		BirthPlace:   getStringValue(req.BirthPlace, current.BirthPlace),
		BirthPlaceID: req.BirthPlaceID,
		DeathDate:    req.DeathDate,
		DeathPlace:   getStringValue(req.DeathPlace, current.DeathPlace),
		BurialPlace:  getStringValue(req.BurialPlace, current.BurialPlace),
		DeathPlaceID: req.DeathPlaceID,
		Occupation:   getStringValue(req.Occupation, current.Occupation),
		Notes:        getStringValue(req.Notes, current.Notes),
		PhotoURL:     req.PhotoURL,
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
	}

	return s.repo.UpdateIndividual(ctx, id, individual)
}

// validateNoCircularRelationship 验证不存在循环关系
func (s *IndividualService) validateNoCircularRelationship(ctx context.Context, childID, parentID int, parentType string) error {
	visited := make(map[int]bool)
	
	var checkAncestors func(int) error
	checkAncestors = func(currentID int) error {
		if visited[currentID] {
			return fmt.Errorf("检测到循环关系：不能将此人设为%s，因为会形成循环父母关系", parentType)
		}
		
		if currentID == childID {
			return fmt.Errorf("检测到循环关系：不能将此人设为%s，因为会形成循环父母关系", parentType)
		}
		
		visited[currentID] = true
		
		// 获取当前人的父母
		individual, err := s.repo.GetIndividualByID(ctx, currentID)
		if err != nil {
			return nil // 如果获取失败，忽略（可能是数据不存在）
		}
		
		// 递归检查父亲
		if individual.FatherID != nil {
			if err := checkAncestors(*individual.FatherID); err != nil {
				return err
			}
		}
		
		// 递归检查母亲
		if individual.MotherID != nil {
			if err := checkAncestors(*individual.MotherID); err != nil {
				return err
			}
		}
		
		return nil
	}
	
	return checkAncestors(parentID)
}

// 辅助函数：获取字符串指针的值或默认值
func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// 辅助函数：获取字符串指针的值或默认值（用于 *string 类型）
func getStringPointerValue(ptr *string, defaultValue *string) *string {
	if ptr != nil {
		return ptr
	}
	return defaultValue
}

// 辅助函数：获取性别指针的值或默认值
func getGenderValue(ptr *models.Gender, defaultValue models.Gender) models.Gender {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// Delete 删除个人信息
func (s *IndividualService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的个人ID")
	}

	// 检查个人是否存在
	_, err := s.repo.GetIndividualByID(ctx, id)
	if err != nil {
		return fmt.Errorf("个人信息不存在")
	}

	// 检查是否有子女
	children, err := s.repo.GetIndividualsByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子女关系失败: %v", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("该个人有子女记录，不能删除。请先删除或转移其子女关系")
	}

	// 检查是否作为配偶存在于家庭关系中
	families, err := s.familyRepo.GetFamiliesByIndividualID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查家庭关系失败: %v", err)
	}
	if len(families) > 0 {
		return fmt.Errorf("该个人仍存在于家庭关系中，不能删除。请先删除相关的家庭关系")
	}

	// 检查其他人是否将此人设为父亲或母亲
	// 通过查询所有个人信息来检查father_id和mother_id字段
	allIndividuals, _, err := s.repo.SearchIndividuals(ctx, "", 10000, 0) // 获取大量数据来检查
	if err != nil {
		return fmt.Errorf("检查父母关系失败: %v", err)
	}
	
	for _, person := range allIndividuals {
		if (person.FatherID != nil && *person.FatherID == id) ||
		   (person.MotherID != nil && *person.MotherID == id) {
			return fmt.Errorf("该个人被其他人设置为父亲或母亲，不能删除。请先调整相关的父母关系")
		}
	}

	// 清理该人作为子女的记录（从子女关系表中删除）
	for _, family := range families {
		err := s.familyRepo.DeleteChild(ctx, family.FamilyID, id)
		if err != nil {
			// 如果删除失败，记录但不阻止整个删除操作
			fmt.Printf("警告：清理子女关系记录失败: %v\n", err)
		}
	}

	// 执行删除
	return s.repo.DeleteIndividual(ctx, id)
}

// Search 搜索个人信息
func (s *IndividualService) Search(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.SearchIndividuals(ctx, query, limit, offset)
}

// GetChildren 获取个人的所有子女
func (s *IndividualService) GetChildren(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}
	return s.repo.GetIndividualsByParentID(ctx, id)
}

// GetParents 获取个人的父母
func (s *IndividualService) GetParents(ctx context.Context, id int) (father, mother *models.Individual, err error) {
	if id <= 0 {
		return nil, nil, fmt.Errorf("无效的个人ID")
	}

	individual, err := s.repo.GetIndividualByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if individual.FatherID != nil {
		father, _ = s.repo.GetIndividualByID(ctx, *individual.FatherID)
	}

	if individual.MotherID != nil {
		mother, _ = s.repo.GetIndividualByID(ctx, *individual.MotherID)
	}

	return father, mother, nil
}

// GetSiblings 获取个人的兄弟姐妹
func (s *IndividualService) GetSiblings(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}

	individual, err := s.repo.GetIndividualByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if individual.FatherID == nil && individual.MotherID == nil {
		return []models.Individual{}, nil
	}

	var siblings []models.Individual

	// 获取同父兄弟姐妹
	if individual.FatherID != nil {
		children, err := s.repo.GetIndividualsByParentID(ctx, *individual.FatherID)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if child.IndividualID != id {
				siblings = append(siblings, child)
			}
		}
	}

	// 获取同母兄弟姐妹（去重）
	if individual.MotherID != nil {
		children, err := s.repo.GetIndividualsByParentID(ctx, *individual.MotherID)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if child.IndividualID != id {
				// 检查是否已存在
				exists := false
				for _, sibling := range siblings {
					if sibling.IndividualID == child.IndividualID {
						exists = true
						break
					}
				}
				if !exists {
					siblings = append(siblings, child)
				}
			}
		}
	}

	return siblings, nil
}

// GetSpouses 获取个人的配偶
func (s *IndividualService) GetSpouses(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}

	return s.repo.GetSpouses(ctx, id)
}

// GetAncestors 获取个人的所有祖先
func (s *IndividualService) GetAncestors(ctx context.Context, id int, generations int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}
	if generations <= 0 {
		generations = 5 // 默认5代
	}
	if generations > 10 {
		generations = 10 // 最多10代
	}

	var ancestors []models.Individual
	visited := make(map[int]bool)

	var getAncestorsRecursive func(int, int) error
	getAncestorsRecursive = func(currentID int, gen int) error {
		if gen <= 0 || visited[currentID] {
			return nil
		}

		visited[currentID] = true
		father, mother, err := s.GetParents(ctx, currentID)
		if err != nil {
			return err
		}

		if father != nil {
			ancestors = append(ancestors, *father)
			if err := getAncestorsRecursive(father.IndividualID, gen-1); err != nil {
				return err
			}
		}

		if mother != nil {
			ancestors = append(ancestors, *mother)
			if err := getAncestorsRecursive(mother.IndividualID, gen-1); err != nil {
				return err
			}
		}

		return nil
	}

	err := getAncestorsRecursive(id, generations)
	return ancestors, err
}

// GetDescendants 获取个人的所有后代
func (s *IndividualService) GetDescendants(ctx context.Context, id int, generations int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}
	if generations <= 0 {
		generations = 5 // 默认5代
	}
	if generations > 10 {
		generations = 10 // 最多10代
	}

	var descendants []models.Individual
	visited := make(map[int]bool)

	var getDescendantsRecursive func(int, int) error
	getDescendantsRecursive = func(currentID int, gen int) error {
		if gen <= 0 || visited[currentID] {
			return nil
		}

		visited[currentID] = true
		children, err := s.repo.GetIndividualsByParentID(ctx, currentID)
		if err != nil {
			return err
		}

		for _, child := range children {
			descendants = append(descendants, child)
			if err := getDescendantsRecursive(child.IndividualID, gen-1); err != nil {
				return err
			}
		}

		return nil
	}

	err := getDescendantsRecursive(id, generations)
	return descendants, err
}

// GetFamilyTree 获取家族树
func (s *IndividualService) GetFamilyTree(ctx context.Context, rootID int, generations int) (*models.FamilyTreeNode, error) {
	if rootID <= 0 {
		return nil, fmt.Errorf("无效的根节点ID")
	}
	if generations <= 0 {
		generations = 3 // 默认3代
	}
	if generations > 10 {
		generations = 10 // 最多10代，防止无限递归
	}

	individual, err := s.repo.GetIndividualByID(ctx, rootID)
	if err != nil {
		return nil, err
	}

	node := &models.FamilyTreeNode{
		Individual: individual,
	}

	if generations > 0 {
		children, err := s.repo.GetIndividualsByParentID(ctx, rootID)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			childNode, err := s.GetFamilyTree(ctx, child.IndividualID, generations-1)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, *childNode)
		}
	}

	return node, nil
}

// AddParent 向上添加父母
func (s *IndividualService) AddParent(ctx context.Context, childID int, req *models.AddParentRequest) (*models.Individual, error) {
	if childID <= 0 {
		return nil, fmt.Errorf("无效的子女ID")
	}

	if req.FullName == "" {
		return nil, fmt.Errorf("父母姓名不能为空")
	}

	// 验证父母类型
	if req.ParentType != "father" && req.ParentType != "mother" {
		return nil, fmt.Errorf("父母类型必须是 'father' 或 'mother'")
	}

	// 获取子女信息
	child, err := s.repo.GetIndividualByID(ctx, childID)
	if err != nil {
		return nil, fmt.Errorf("获取子女信息失败: %v", err)
	}

	// 检查是否已经有对应的父母
	if req.ParentType == "father" && child.FatherID != nil {
		return nil, fmt.Errorf("该子女已经有父亲了")
	}
	if req.ParentType == "mother" && child.MotherID != nil {
		return nil, fmt.Errorf("该子女已经有母亲了")
	}

	// 根据父母类型设置性别
	if req.ParentType == "father" {
		req.Gender = models.GenderMale
	} else {
		req.Gender = models.GenderFemale
	}

	// 创建父母个人信息
	parent := &models.Individual{
		FullName:     req.FullName,
		Gender:       req.Gender,
		BirthDate:    req.BirthDate,
		BirthPlace:   req.BirthPlace,
		BirthPlaceID: req.BirthPlaceID,
		DeathDate:    req.DeathDate,
		DeathPlace:   req.DeathPlace,
		DeathPlaceID: req.DeathPlaceID,
		Occupation:   req.Occupation,
		Notes:        req.Notes,
		PhotoURL:     req.PhotoURL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 创建父母记录
	createdParent, err := s.repo.CreateIndividual(ctx, parent)
	if err != nil {
		return nil, fmt.Errorf("创建父母记录失败: %v", err)
	}

	// 获取所有兄弟姐妹（包括当前子女）
	siblings, err := s.findSiblingsForParentUpdate(ctx, childID)
	if err != nil {
		// 如果获取兄弟姐妹失败，至少更新当前子女
		siblings = []models.Individual{*child}
		fmt.Printf("警告: 获取兄弟姐妹失败，只更新当前子女: %v\n", err)
	}

	// 更新所有兄弟姐妹的父母关系
	var failedUpdates []string
	for _, sibling := range siblings {
		updateReq := &models.UpdateIndividualRequest{
			FullName:     &sibling.FullName,
			Gender:       &sibling.Gender,
			BirthDate:    sibling.BirthDate,
			BirthPlace:   &sibling.BirthPlace,
			BirthPlaceID: sibling.BirthPlaceID,
			DeathDate:    sibling.DeathDate,
			DeathPlace:   &sibling.DeathPlace,
			DeathPlaceID: sibling.DeathPlaceID,
			Occupation:   &sibling.Occupation,
			Notes:        &sibling.Notes,
			PhotoURL:     sibling.PhotoURL,
			FatherID:     sibling.FatherID,
			MotherID:     sibling.MotherID,
		}

		if req.ParentType == "father" {
			updateReq.FatherID = &createdParent.IndividualID
		} else {
			updateReq.MotherID = &createdParent.IndividualID
		}

		_, err = s.Update(ctx, sibling.IndividualID, updateReq)
		if err != nil {
			failedUpdates = append(failedUpdates, fmt.Sprintf("%s(ID:%d)", sibling.FullName, sibling.IndividualID))
			fmt.Printf("警告: 更新%s的父母关系失败: %v\n", sibling.FullName, err)
		}
	}

	// 如果有更新失败的情况，记录但不回滚
	if len(failedUpdates) > 0 {
		fmt.Printf("警告: 以下成员的父母关系更新失败: %v\n", failedUpdates)
	}

	// 检查并创建父母之间的夫妻关系
	err = s.ensureParentsMarriageForAllSiblings(ctx, siblings)
	if err != nil {
		// 记录警告但不影响主要操作
		fmt.Printf("警告: 创建父母夫妻关系失败: %v\n", err)
	}

	return createdParent, nil
}

// findSiblingsForParentUpdate 查找需要更新父母关系的所有兄弟姐妹
func (s *IndividualService) findSiblingsForParentUpdate(ctx context.Context, childID int) ([]models.Individual, error) {
	child, err := s.repo.GetIndividualByID(ctx, childID)
	if err != nil {
		return nil, err
	}

	var allSiblings []models.Individual
	siblingMap := make(map[int]bool) // 用于去重

	// 添加当前子女
	allSiblings = append(allSiblings, *child)
	siblingMap[child.IndividualID] = true

	// 如果有父亲，获取所有同父兄弟姐妹
	if child.FatherID != nil {
		fatherChildren, err := s.repo.GetIndividualsByParentID(ctx, *child.FatherID)
		if err == nil {
			for _, sibling := range fatherChildren {
				if !siblingMap[sibling.IndividualID] {
					allSiblings = append(allSiblings, sibling)
					siblingMap[sibling.IndividualID] = true
				}
			}
		}
	}

	// 如果有母亲，获取所有同母兄弟姐妹
	if child.MotherID != nil {
		motherChildren, err := s.repo.GetIndividualsByParentID(ctx, *child.MotherID)
		if err == nil {
			for _, sibling := range motherChildren {
				if !siblingMap[sibling.IndividualID] {
					allSiblings = append(allSiblings, sibling)
					siblingMap[sibling.IndividualID] = true
				}
			}
		}
	}

	return allSiblings, nil
}

// ensureParentsMarriageForAllSiblings 确保所有兄弟姐妹的父母之间有夫妻关系
func (s *IndividualService) ensureParentsMarriageForAllSiblings(ctx context.Context, siblings []models.Individual) error {
	if len(siblings) == 0 {
		return nil
	}

	// 获取第一个兄弟姐妹的父母信息作为参考
	var fatherID, motherID *int
	for _, sibling := range siblings {
		if sibling.FatherID != nil && sibling.MotherID != nil {
			fatherID = sibling.FatherID
			motherID = sibling.MotherID
			break
		}
	}

	// 如果没有找到完整的父母信息，尝试从所有兄弟姐妹中收集
	if fatherID == nil || motherID == nil {
		for _, sibling := range siblings {
			if fatherID == nil && sibling.FatherID != nil {
				fatherID = sibling.FatherID
			}
			if motherID == nil && sibling.MotherID != nil {
				motherID = sibling.MotherID
			}
			if fatherID != nil && motherID != nil {
				break
			}
		}
	}

	// 如果仍然没有完整的父母信息，无法创建夫妻关系
	if fatherID == nil || motherID == nil {
		return fmt.Errorf("无法找到完整的父母信息来创建夫妻关系")
	}

	// 检查父母之间是否已经有夫妻关系
	families, err := s.familyRepo.GetFamiliesByIndividualID(ctx, *fatherID)
	if err != nil {
		return fmt.Errorf("检查现有夫妻关系失败: %v", err)
	}

	// 检查是否已存在夫妻关系
	for _, family := range families {
		if family.HusbandID != nil && *family.HusbandID == *fatherID &&
			family.WifeID != nil && *family.WifeID == *motherID {
			// 夫妻关系已存在
			return nil
		}
	}

	// 获取下一个婚姻顺序
	marriageOrder, err := s.getNextMarriageOrder(ctx, *fatherID)
	if err != nil {
		return fmt.Errorf("获取婚姻顺序失败: %v", err)
	}

	// 创建夫妻关系
	familyReq := &models.CreateFamilyRequest{
		HusbandID: fatherID,
		WifeID:    motherID,
	}

	family := &models.Family{
		HusbandID:     familyReq.HusbandID,
		WifeID:        familyReq.WifeID,
		MarriageOrder: marriageOrder,
		MarriageDate:  familyReq.MarriageDate,
		DivorceDate:   familyReq.DivorceDate,
		Notes:         familyReq.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = s.familyRepo.CreateFamily(ctx, family)
	if err != nil {
		return fmt.Errorf("创建夫妻关系失败: %v", err)
	}

	return nil
}

// getNextMarriageOrder 获取下一个婚姻顺序
func (s *IndividualService) getNextMarriageOrder(ctx context.Context, husbandID int) (int, error) {
	families, err := s.familyRepo.GetFamiliesByIndividualID(ctx, husbandID)
	if err != nil {
		return 1, err
	}

	maxOrder := 0
	for _, family := range families {
		if family.HusbandID != nil && *family.HusbandID == husbandID {
			if family.MarriageOrder > maxOrder {
				maxOrder = family.MarriageOrder
			}
		}
	}

	return maxOrder + 1, nil
}
