package services

import (
	"context"
	"fmt"
	"time"

	"familytree/interfaces"
	"familytree/models"
)

// FamilyService 家庭关系服务实现
type FamilyService struct {
	repo           interfaces.FamilyRepository
	individualRepo interfaces.IndividualRepository
}

// NewFamilyService 创建新的家庭关系服务
func NewFamilyService(repo interfaces.FamilyRepository, individualRepo interfaces.IndividualRepository) interfaces.FamilyService {
	return &FamilyService{
		repo:           repo,
		individualRepo: individualRepo,
	}
}

// CreateFamily 创建家庭关系
func (s *FamilyService) CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error) {
	// 验证输入
	if req.HusbandID == nil && req.WifeID == nil {
		return nil, fmt.Errorf("至少需要指定丈夫或妻子")
	}

	// 验证夫妻不能是同一个人
	if req.HusbandID != nil && req.WifeID != nil && *req.HusbandID == *req.WifeID {
		return nil, fmt.Errorf("夫妻不能是同一个人")
	}

	// 验证丈夫性别
	if req.HusbandID != nil {
		husband, err := s.individualRepo.GetIndividualByID(ctx, *req.HusbandID)
		if err != nil {
			return nil, fmt.Errorf("丈夫信息不存在")
		}
		if husband.Gender != models.GenderMale {
			return nil, fmt.Errorf("丈夫必须是男性")
		}
	}

	// 验证妻子性别
	if req.WifeID != nil {
		wife, err := s.individualRepo.GetIndividualByID(ctx, *req.WifeID)
		if err != nil {
			return nil, fmt.Errorf("妻子信息不存在")
		}
		if wife.Gender != models.GenderFemale {
			return nil, fmt.Errorf("妻子必须是女性")
		}
	}

	// 计算婚姻顺序
	var marriageOrder int = 1
	if req.HusbandID != nil {
		// 获取丈夫的现有家庭关系，计算婚姻顺序
		existingFamilies, err := s.repo.GetFamiliesByIndividualID(ctx, *req.HusbandID)
		if err != nil {
			return nil, fmt.Errorf("获取现有家庭关系失败: %v", err)
		}

		for _, family := range existingFamilies {
			if family.HusbandID != nil && *family.HusbandID == *req.HusbandID {
				if family.MarriageOrder >= marriageOrder {
					marriageOrder = family.MarriageOrder + 1
				}
			}
		}
	}

	// 创建家庭记录
	family := &models.Family{
		HusbandID:       req.HusbandID,
		WifeID:          req.WifeID,
		MarriageOrder:   marriageOrder, // 使用计算出的婚姻顺序
		MarriageDate:    req.MarriageDate,
		MarriagePlaceID: req.MarriagePlaceID,
		DivorceDate:     req.DivorceDate,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return s.repo.CreateFamily(ctx, family)
}

// GetByID 根据ID获取家庭关系
func (s *FamilyService) GetByID(ctx context.Context, id int) (*models.Family, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的家庭ID")
	}
	return s.repo.GetFamilyByID(ctx, id)
}

// Update 更新家庭关系
func (s *FamilyService) Update(ctx context.Context, id int, req *models.CreateFamilyRequest) (*models.Family, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的家庭ID")
	}

	// 验证输入
	if req.HusbandID == nil && req.WifeID == nil {
		return nil, fmt.Errorf("至少需要指定丈夫或妻子")
	}

	// 验证夫妻不能是同一个人
	if req.HusbandID != nil && req.WifeID != nil && *req.HusbandID == *req.WifeID {
		return nil, fmt.Errorf("夫妻不能是同一个人")
	}

	// 获取现有家庭记录
	current, err := s.repo.GetFamilyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新家庭记录
	family := &models.Family{
		FamilyID:        id,
		HusbandID:       req.HusbandID,
		WifeID:          req.WifeID,
		MarriageOrder:   current.MarriageOrder, // 保持原有的婚姻顺序
		MarriageDate:    req.MarriageDate,
		MarriagePlaceID: req.MarriagePlaceID,
		DivorceDate:     req.DivorceDate,
		Notes:           req.Notes,
		CreatedAt:       current.CreatedAt,
		UpdatedAt:       time.Now(),
	}

	return s.repo.UpdateFamily(ctx, id, family)
}

// Delete 删除家庭关系
func (s *FamilyService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的家庭ID")
	}

	// 检查家庭关系是否存在
	_, err := s.repo.GetFamilyByID(ctx, id)
	if err != nil {
		return fmt.Errorf("家庭关系不存在")
	}

	// 检查是否有子女记录
	children, err := s.repo.GetChildrenByFamilyID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子女关系失败: %v", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("该家庭有子女记录，不能删除。请先删除或转移子女关系")
	}

	// 先清理所有相关的子女关系记录
	for _, child := range children {
		err := s.repo.DeleteChild(ctx, id, child.IndividualID)
		if err != nil {
			return fmt.Errorf("删除子女关系记录失败: %v", err)
		}
	}

	// 删除家庭关系
	return s.repo.DeleteFamily(ctx, id)
}

// GetBySpouses 根据夫妻ID获取家庭关系
func (s *FamilyService) GetBySpouses(ctx context.Context, husbandID, wifeID int) (*models.Family, error) {
	families, err := s.repo.GetFamiliesByIndividualID(ctx, husbandID)
	if err != nil {
		return nil, err
	}

	for _, family := range families {
		if (family.HusbandID != nil && *family.HusbandID == husbandID &&
			family.WifeID != nil && *family.WifeID == wifeID) ||
			(family.HusbandID != nil && *family.HusbandID == wifeID &&
				family.WifeID != nil && *family.WifeID == husbandID) {
			return &family, nil
		}
	}

	return nil, fmt.Errorf("未找到对应的家庭关系")
}

// GetByIndividualID 获取某人参与的所有家庭关系
func (s *FamilyService) GetByIndividualID(ctx context.Context, individualID int) ([]models.Family, error) {
	if individualID <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}
	return s.repo.GetFamiliesByIndividualID(ctx, individualID)
}

// AddSpouse 添加配偶关系
func (s *FamilyService) AddSpouse(ctx context.Context, individualID, spouseID int) (*models.Family, error) {
	if individualID <= 0 || spouseID <= 0 {
		return nil, fmt.Errorf("无效的个人ID")
	}

	if individualID == spouseID {
		return nil, fmt.Errorf("不能将自己设为配偶")
	}

	// 获取个人信息
	individual, err := s.individualRepo.GetIndividualByID(ctx, individualID)
	if err != nil {
		return nil, fmt.Errorf("个人信息不存在")
	}

	spouse, err := s.individualRepo.GetIndividualByID(ctx, spouseID)
	if err != nil {
		return nil, fmt.Errorf("配偶信息不存在")
	}

	// 检查是否已经存在相同的配偶关系
	existingFamilies, err := s.repo.GetFamiliesByIndividualID(ctx, individualID)
	if err != nil {
		return nil, err
	}

	for _, family := range existingFamilies {
		if (family.HusbandID != nil && *family.HusbandID == spouseID) ||
			(family.WifeID != nil && *family.WifeID == spouseID) {
			return nil, fmt.Errorf("已存在相同的配偶关系")
		}
	}

	// 根据性别确定夫妻角色并计算婚姻顺序
	var req *models.CreateFamilyRequest
	var marriageOrder int = 1

	if individual.Gender == models.GenderMale && spouse.Gender == models.GenderFemale {
		// 男性添加妻子 - 计算他的妻子数量
		for _, family := range existingFamilies {
			if family.HusbandID != nil && *family.HusbandID == individualID {
				if family.MarriageOrder >= marriageOrder {
					marriageOrder = family.MarriageOrder + 1
				}
			}
		}
		req = &models.CreateFamilyRequest{
			HusbandID: &individualID,
			WifeID:    &spouseID,
		}
	} else if individual.Gender == models.GenderFemale && spouse.Gender == models.GenderMale {
		// 女性添加丈夫 - 计算丈夫的妻子数量
		spouseFamilies, err := s.repo.GetFamiliesByIndividualID(ctx, spouseID)
		if err != nil {
			return nil, err
		}
		for _, family := range spouseFamilies {
			if family.HusbandID != nil && *family.HusbandID == spouseID {
				if family.MarriageOrder >= marriageOrder {
					marriageOrder = family.MarriageOrder + 1
				}
			}
		}
		req = &models.CreateFamilyRequest{
			HusbandID: &spouseID,
			WifeID:    &individualID,
		}
	} else {
		return nil, fmt.Errorf("配偶关系必须是一男一女")
	}

	// 创建家庭关系，包含婚姻顺序
	family := &models.Family{
		HusbandID:       req.HusbandID,
		WifeID:          req.WifeID,
		MarriageOrder:   marriageOrder,
		MarriageDate:    req.MarriageDate,
		MarriagePlaceID: req.MarriagePlaceID,
		DivorceDate:     req.DivorceDate,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return s.repo.CreateFamily(ctx, family)
}

// AddChild 为家庭添加子女
func (s *FamilyService) AddChild(ctx context.Context, familyID, childID int, relationship string) error {
	if familyID <= 0 || childID <= 0 {
		return fmt.Errorf("无效的ID参数")
	}

	// 验证家庭存在
	_, err := s.repo.GetFamilyByID(ctx, familyID)
	if err != nil {
		return fmt.Errorf("家庭关系不存在")
	}

	// 验证子女存在
	_, err = s.individualRepo.GetIndividualByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("子女信息不存在")
	}

	// 创建子女关系记录
	child := &models.Child{
		FamilyID:              familyID,
		IndividualID:          childID,
		RelationshipToParents: relationship,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	_, err = s.repo.CreateChild(ctx, child)
	return err
}

// RemoveChild 从家庭移除子女
func (s *FamilyService) RemoveChild(ctx context.Context, familyID, childID int) error {
	if familyID <= 0 || childID <= 0 {
		return fmt.Errorf("无效的ID参数")
	}
	return s.repo.DeleteChild(ctx, familyID, childID)
}

// GetChildren 获取家庭的所有子女
func (s *FamilyService) GetChildren(ctx context.Context, familyID int) ([]models.Child, error) {
	if familyID <= 0 {
		return nil, fmt.Errorf("无效的家庭ID")
	}
	return s.repo.GetChildrenByFamilyID(ctx, familyID)
}
