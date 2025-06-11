package services

import (
	"context"
	"fmt"

	"familytree/interfaces"
	"familytree/models"
)

// IndividualService 个人服务实现
type IndividualService struct {
	repo interfaces.IndividualRepository
}

// NewIndividualService 创建新的个人服务
func NewIndividualService(repo interfaces.IndividualRepository) interfaces.IndividualService {
	return &IndividualService{repo: repo}
}

// Create 创建个人信息
func (s *IndividualService) Create(ctx context.Context, req *models.CreateIndividualRequest) (*models.Individual, error) {
	// 验证输入
	if req.FullName == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}

	// 验证父母关系
	if req.FatherID != nil && req.MotherID != nil && *req.FatherID == *req.MotherID {
		return nil, fmt.Errorf("父亲和母亲不能是同一个人")
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

	// 创建个人信息
	individual := &models.Individual{
		FullName:     req.FullName,
		Gender:       req.Gender,
		BirthDate:    req.BirthDate,
		BirthPlaceID: req.BirthPlaceID,
		DeathDate:    req.DeathDate,
		DeathPlaceID: req.DeathPlaceID,
		Occupation:   req.Occupation,
		Notes:        req.Notes,
		PhotoURL:     req.PhotoURL,
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
	}

	return s.repo.CreateIndividual(ctx, individual)
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
	if req.FullName == nil || *req.FullName == "" {
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
				BirthPlaceID: child.BirthPlaceID,
				DeathDate:    child.DeathDate,
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
		req.PhotoURL = &photoURL
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
		PhotoURL:     getStringValue(req.PhotoURL, current.PhotoURL),
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
	}

	return s.repo.UpdateIndividual(ctx, id, individual)
}

// 辅助函数：获取字符串指针的值或默认值
func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
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

	// 检查是否有子女
	children, err := s.repo.GetIndividualsByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子女关系失败: %v", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("该个人有子女记录，不能删除")
	}

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
