package services

import (
	"context"
	"fmt"

	"familytree/interfaces"
	"familytree/models"
	"familytree/pkg/middleware"
)

// UserService 用户服务实现
type UserService struct {
	userRepo interfaces.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo interfaces.UserRepository) interfaces.UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(ctx context.Context, id int) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

// Update 更新用户信息
func (s *UserService) Update(ctx context.Context, id int, user *models.User) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	// 验证输入
	if user.Username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	if user.Email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	if user.FullName == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}

	// 检查用户名是否被其他用户使用
	existingUser, _ := s.userRepo.GetUserByUsername(ctx, user.Username)
	if existingUser != nil && existingUser.UserID != id {
		return nil, fmt.Errorf("用户名已被其他用户使用")
	}

	// 检查邮箱是否被其他用户使用
	existingEmail, _ := s.userRepo.GetUserByEmail(ctx, user.Email)
	if existingEmail != nil && existingEmail.UserID != id {
		return nil, fmt.Errorf("邮箱已被其他用户使用")
	}

	updatedUser, err := s.userRepo.UpdateUser(ctx, id, user)
	if err != nil {
		return nil, err
	}

	// 清除密码字段
	updatedUser.Password = ""
	return updatedUser, nil
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的用户ID")
	}

	// 检查用户是否存在
	_, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: 在删除用户之前，可以添加清理相关数据的逻辑
	// 例如：删除用户的家族树数据等

	return s.userRepo.DeleteUser(ctx, id)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	if userID <= 0 {
		return fmt.Errorf("无效的用户ID")
	}

	if oldPassword == "" {
		return fmt.Errorf("旧密码不能为空")
	}

	if newPassword == "" {
		return fmt.Errorf("新密码不能为空")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("新密码长度至少6个字符")
	}

	// 获取用户信息（包含密码）
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if !middleware.CheckPassword(user.Password, oldPassword) {
		return fmt.Errorf("旧密码错误")
	}

	// 哈希新密码
	hashedPassword, err := middleware.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("新密码处理失败: %v", err)
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}

// FamilyTreeService 家族树服务实现
type FamilyTreeService struct {
	familyTreeRepo    interfaces.FamilyTreeRepository
	individualRepo    interfaces.IndividualRepository
	individualService interfaces.IndividualService
}

// NewFamilyTreeService 创建家族树服务
func NewFamilyTreeService(familyTreeRepo interfaces.FamilyTreeRepository, individualRepo interfaces.IndividualRepository, individualService interfaces.IndividualService) interfaces.FamilyTreeService {
	return &FamilyTreeService{
		familyTreeRepo:    familyTreeRepo,
		individualRepo:    individualRepo,
		individualService: individualService,
	}
}

// CreateFamilyTree 创建家族树
func (s *FamilyTreeService) CreateFamilyTree(ctx context.Context, userID int, req *models.CreateFamilyTreeRequest) (*models.UserFamilyTree, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	if req.FamilyTreeName == "" {
		return nil, fmt.Errorf("家族树名称不能为空")
	}

	// 创建家族树
	familyTree := &models.UserFamilyTree{
		UserID:         userID,
		FamilyTreeName: req.FamilyTreeName,
		Description:    req.Description,
		IsDefault:      false, // 新创建的家族树默认不是默认家族树
	}

	createdFamilyTree, err := s.familyTreeRepo.CreateFamilyTree(ctx, familyTree)
	if err != nil {
		return nil, err
	}

	// 如果提供了根人员信息，创建根人员
	if req.RootPersonInfo != nil {
		// 添加用户ID和家族树ID到个人信息
		req.RootPersonInfo.FatherID = nil // 根人员没有父亲
		req.RootPersonInfo.MotherID = nil // 根人员没有母亲

		rootPerson, err := s.individualService.Create(ctx, req.RootPersonInfo)
		if err != nil {
			// 如果创建根人员失败，不删除家族树，只记录错误
			fmt.Printf("创建根人员失败: %v\n", err)
		} else {
			// 更新家族树的根人员ID
			createdFamilyTree.RootPersonID = &rootPerson.IndividualID
			updatedFamilyTree, err := s.familyTreeRepo.UpdateFamilyTree(ctx, createdFamilyTree.FamilyTreeID, createdFamilyTree)
			if err != nil {
				fmt.Printf("更新家族树根人员失败: %v\n", err)
			} else {
				createdFamilyTree = updatedFamilyTree
			}
		}
	}

	return createdFamilyTree, nil
}

// GetUserFamilyTrees 获取用户的家族树列表
func (s *FamilyTreeService) GetUserFamilyTrees(ctx context.Context, userID int) ([]models.UserFamilyTree, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	return s.familyTreeRepo.GetUserFamilyTrees(ctx, userID)
}

// GetDefaultFamilyTree 获取默认家族树
func (s *FamilyTreeService) GetDefaultFamilyTree(ctx context.Context, userID int) (*models.UserFamilyTree, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	return s.familyTreeRepo.GetDefaultFamilyTree(ctx, userID)
}

// SetDefaultFamilyTree 设置默认家族树
func (s *FamilyTreeService) SetDefaultFamilyTree(ctx context.Context, userID int, familyTreeID int) error {
	if userID <= 0 {
		return fmt.Errorf("无效的用户ID")
	}

	if familyTreeID <= 0 {
		return fmt.Errorf("无效的家族树ID")
	}

	// 验证家族树是否属于该用户
	familyTree, err := s.familyTreeRepo.GetFamilyTreeByID(ctx, familyTreeID)
	if err != nil {
		return err
	}

	if familyTree.UserID != userID {
		return fmt.Errorf("家族树不属于该用户")
	}

	return s.familyTreeRepo.SetDefaultFamilyTree(ctx, userID, familyTreeID)
}

// DeleteFamilyTree 删除家族树
func (s *FamilyTreeService) DeleteFamilyTree(ctx context.Context, userID int, familyTreeID int) error {
	if userID <= 0 {
		return fmt.Errorf("无效的用户ID")
	}

	if familyTreeID <= 0 {
		return fmt.Errorf("无效的家族树ID")
	}

	// 验证家族树是否属于该用户
	familyTree, err := s.familyTreeRepo.GetFamilyTreeByID(ctx, familyTreeID)
	if err != nil {
		return err
	}

	if familyTree.UserID != userID {
		return fmt.Errorf("家族树不属于该用户")
	}

	// 检查是否是唯一的家族树
	userFamilyTrees, err := s.familyTreeRepo.GetUserFamilyTrees(ctx, userID)
	if err != nil {
		return fmt.Errorf("检查用户家族树失败: %v", err)
	}

	if len(userFamilyTrees) <= 1 {
		return fmt.Errorf("不能删除唯一的家族树")
	}

	// TODO: 在删除家族树之前，可以添加清理相关数据的逻辑
	// 例如：删除该家族树下的所有个人信息、家庭关系等

	return s.familyTreeRepo.DeleteFamilyTree(ctx, familyTreeID)
}

// UpdateFamilyTree 更新家族树信息
func (s *FamilyTreeService) UpdateFamilyTree(ctx context.Context, userID int, familyTreeID int, req *models.CreateFamilyTreeRequest) (*models.UserFamilyTree, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	if familyTreeID <= 0 {
		return nil, fmt.Errorf("无效的家族树ID")
	}

	if req.FamilyTreeName == "" {
		return nil, fmt.Errorf("家族树名称不能为空")
	}

	// 验证家族树是否属于该用户
	familyTree, err := s.familyTreeRepo.GetFamilyTreeByID(ctx, familyTreeID)
	if err != nil {
		return nil, err
	}

	if familyTree.UserID != userID {
		return nil, fmt.Errorf("家族树不属于该用户")
	}

	// 更新家族树信息
	familyTree.FamilyTreeName = req.FamilyTreeName
	familyTree.Description = req.Description

	return s.familyTreeRepo.UpdateFamilyTree(ctx, familyTreeID, familyTree)
}
