package services

import (
	"context"
	"log"

	"familytree/interfaces"
	"familytree/models"
	"familytree/pkg/errors"
	"familytree/pkg/objectpool"
	"familytree/repository"
)

// CachedIndividualService 带缓存的个人信息服务
type CachedIndividualService struct {
	service    interfaces.IndividualService
	cache      *repository.CacheRepository
	objectPool *objectpool.IndividualPool
	treePool   *objectpool.FamilyTreeNodePool
}

// NewCachedIndividualService 创建带缓存的个人信息服务
func NewCachedIndividualService(
	service interfaces.IndividualService,
	cache *repository.CacheRepository,
) interfaces.IndividualService {
	return &CachedIndividualService{
		service:    service,
		cache:      cache,
		objectPool: objectpool.NewIndividualPool(),
		treePool:   objectpool.NewFamilyTreeNodePool(),
	}
}

// Create 创建个人信息（创建后清除相关缓存）
func (s *CachedIndividualService) Create(ctx context.Context, req *models.CreateIndividualRequest) (*models.Individual, error) {
	// 验证输入
	if req.FullName == "" {
		return nil, errors.New(errors.ErrCodeInvalidInput, "姓名不能为空")
	}

	// 调用原服务
	individual, err := s.service.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 异步清除相关缓存
	go func() {
		s.invalidateRelatedCache(context.Background(), individual.IndividualID)
	}()

	return individual, nil
}

// CreateForUser 创建个人信息（用户隔离版本，创建后清除相关缓存）
func (s *CachedIndividualService) CreateForUser(ctx context.Context, userID int, req *models.CreateIndividualRequest) (*models.Individual, error) {
	// 验证输入
	if req.FullName == "" {
		return nil, errors.New(errors.ErrCodeInvalidInput, "姓名不能为空")
	}

	// 调用原服务
	individual, err := s.service.CreateForUser(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// 异步清除相关缓存
	go func() {
		s.invalidateRelatedCache(context.Background(), individual.IndividualID)
	}()

	return individual, nil
}

// GetByID 获取个人信息（带缓存）
func (s *CachedIndividualService) GetByID(ctx context.Context, id int) (*models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	// 尝试从缓存获取
	if s.cache != nil {
		cached, err := s.cache.GetIndividual(ctx, id)
		if err == nil && cached != nil {
			log.Printf("缓存命中：个人信息 ID=%d", id)
			return cached, nil
		}
	}

	// 缓存未命中，从数据库获取
	individual, err := s.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 异步写入缓存
	if s.cache != nil {
		go func() {
			if err := s.cache.SetIndividual(context.Background(), individual); err != nil {
				log.Printf("写入缓存失败：%v", err)
			}
		}()
	}

	return individual, nil
}

// Update 更新个人信息（更新后清除相关缓存）
func (s *CachedIndividualService) Update(ctx context.Context, id int, req *models.UpdateIndividualRequest) (*models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	// 调用原服务
	individual, err := s.service.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// 异步清除相关缓存
	go func() {
		s.invalidateRelatedCache(context.Background(), id)
	}()

	return individual, nil
}

// Delete 删除个人信息（删除后清除相关缓存）
func (s *CachedIndividualService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.ErrInvalidID
	}

	// 调用原服务
	err := s.service.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 异步清除相关缓存
	go func() {
		s.invalidateRelatedCache(context.Background(), id)
	}()

	return nil
}

// Search 搜索个人信息（不缓存搜索结果，因为变化太频繁）
func (s *CachedIndividualService) Search(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	return s.service.Search(ctx, query, limit, offset)
}

// SearchForUser 搜索个人信息（用户隔离版本，不缓存搜索结果，因为变化太频繁）
func (s *CachedIndividualService) SearchForUser(ctx context.Context, userID int, query string, limit, offset int) ([]models.Individual, int, error) {
	return s.service.SearchForUser(ctx, userID, query, limit, offset)
}

// GetChildren 获取子女（带缓存）
func (s *CachedIndividualService) GetChildren(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	return s.service.GetChildren(ctx, id)
}

// GetParents 获取父母（带缓存）
func (s *CachedIndividualService) GetParents(ctx context.Context, id int) (*models.Individual, *models.Individual, error) {
	if id <= 0 {
		return nil, nil, errors.ErrInvalidID
	}

	return s.service.GetParents(ctx, id)
}

// GetSiblings 获取兄弟姐妹
func (s *CachedIndividualService) GetSiblings(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	return s.service.GetSiblings(ctx, id)
}

// GetSpouses 获取配偶
func (s *CachedIndividualService) GetSpouses(ctx context.Context, id int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	return s.service.GetSpouses(ctx, id)
}

// GetAncestors 获取祖先
func (s *CachedIndividualService) GetAncestors(ctx context.Context, id int, generations int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	return s.service.GetAncestors(ctx, id, generations)
}

// GetDescendants 获取后代
func (s *CachedIndividualService) GetDescendants(ctx context.Context, id int, generations int) ([]models.Individual, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidID
	}

	return s.service.GetDescendants(ctx, id, generations)
}

// GetFamilyTree 获取家族树（带缓存）
func (s *CachedIndividualService) GetFamilyTree(ctx context.Context, rootID int, generations int) (*models.FamilyTreeNode, error) {
	if rootID <= 0 {
		return nil, errors.ErrInvalidID
	}

	// 尝试从缓存获取家族树
	if s.cache != nil {
		cached, err := s.cache.GetFamilyTree(ctx, rootID)
		if err == nil && cached != nil {
			log.Printf("缓存命中：家族树 RootID=%d", rootID)
			return cached, nil
		}
	}

	// 缓存未命中，从服务获取
	tree, err := s.service.GetFamilyTree(ctx, rootID, generations)
	if err != nil {
		return nil, err
	}

	// 异步写入缓存
	if s.cache != nil {
		go func() {
			if err := s.cache.SetFamilyTree(context.Background(), rootID, tree); err != nil {
				log.Printf("写入家族树缓存失败：%v", err)
			}
		}()
	}

	return tree, nil
}

// AddParent 添加父母
func (s *CachedIndividualService) AddParent(ctx context.Context, childID int, req *models.AddParentRequest) (*models.Individual, error) {
	if childID <= 0 {
		return nil, errors.ErrInvalidID
	}

	parent, err := s.service.AddParent(ctx, childID, req)
	if err != nil {
		return nil, err
	}

	// 异步清除相关缓存
	go func() {
		s.invalidateRelatedCache(context.Background(), childID)
		if parent != nil {
			s.invalidateRelatedCache(context.Background(), parent.IndividualID)
		}
	}()

	return parent, nil
}

// invalidateRelatedCache 清除相关缓存
func (s *CachedIndividualService) invalidateRelatedCache(ctx context.Context, id int) {
	if s.cache == nil {
		return
	}

	// 清除个人缓存
	if err := s.cache.DeleteIndividual(ctx, id); err != nil {
		log.Printf("清除个人缓存失败 ID=%d: %v", id, err)
	}

	// 清除家族树缓存（以此人为根节点）
	if err := s.cache.DeleteFamilyTree(ctx, id); err != nil {
		log.Printf("清除家族树缓存失败 RootID=%d: %v", id, err)
	}

	// TODO: 可以根据需要清除更多相关缓存
	log.Printf("已清除个人 ID=%d 的相关缓存", id)
}
