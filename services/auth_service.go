package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"familytree/interfaces"
	"familytree/models"
	"familytree/pkg/middleware"
)

// AuthService 认证服务实现
type AuthService struct {
	userRepo       interfaces.UserRepository
	familyTreeRepo interfaces.FamilyTreeRepository
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo interfaces.UserRepository, familyTreeRepo interfaces.FamilyTreeRepository) interfaces.AuthService {
	return &AuthService{
		userRepo:       userRepo,
		familyTreeRepo: familyTreeRepo,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// 验证输入
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.GetUserByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingEmail, _ := s.userRepo.GetUserByEmail(ctx, req.Email)
	if existingEmail != nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 哈希密码
	hashedPassword, err := middleware.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码处理失败: %v", err)
	}

	// 创建用户
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FullName:  req.FullName,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 为新用户创建默认家族树
	defaultFamilyTree := &models.UserFamilyTree{
		UserID:         createdUser.UserID,
		FamilyTreeName: fmt.Sprintf("%s的家族树", createdUser.FullName),
		Description:    "我的第一个家族树",
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = s.familyTreeRepo.CreateFamilyTree(ctx, defaultFamilyTree)
	if err != nil {
		// 日志记录错误，但不影响用户注册
		fmt.Printf("创建默认家族树失败: %v\n", err)
	}

	// 清除密码字段（安全考虑）
	createdUser.Password = ""
	return createdUser, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// 验证输入
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// 查找用户（支持用户名或邮箱登录）
	var user *models.User
	var err error

	if s.isEmail(req.Username) {
		user, err = s.userRepo.GetUserByEmail(ctx, req.Username)
	} else {
		user, err = s.userRepo.GetUserByUsername(ctx, req.Username)
	}

	if err != nil || user == nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, fmt.Errorf("用户账户已被禁用")
	}

	// 验证密码
	if !middleware.CheckPassword(user.Password, req.Password) {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	// 清除密码字段
	user.Password = ""

	return &models.LoginResponse{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 60 * 60, // 24小时，以秒为单位
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// 验证刷新令牌
	claims, err := middleware.ValidateJWTToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("无效的刷新令牌")
	}

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户是否仍然激活
	if !user.IsActive {
		return nil, fmt.Errorf("用户账户已被禁用")
	}

	// 生成新的令牌
	newAccessToken, newRefreshToken, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成新令牌失败: %v", err)
	}

	// 清除密码字段
	user.Password = ""

	return &models.LoginResponse{
		User:         user,
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    24 * 60 * 60,
	}, nil
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := middleware.ValidateJWTToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("用户账户已被禁用")
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

// GenerateToken 生成JWT令牌
func (s *AuthService) GenerateToken(user *models.User) (string, string, error) {
	return middleware.GenerateJWTToken(user)
}

// validateRegisterRequest 验证注册请求
func (s *AuthService) validateRegisterRequest(req *models.RegisterRequest) error {
	if req.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	if len(req.Username) < 3 || len(req.Username) > 32 {
		return fmt.Errorf("用户名长度必须在3-32个字符之间")
	}

	// 用户名只能包含字母、数字、下划线
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Username) {
		return fmt.Errorf("用户名只能包含字母、数字和下划线")
	}

	if req.Email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	if !s.isEmail(req.Email) {
		return fmt.Errorf("邮箱格式无效")
	}

	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if len(req.Password) < 6 {
		return fmt.Errorf("密码长度至少6个字符")
	}

	if req.FullName == "" {
		return fmt.Errorf("姓名不能为空")
	}

	if len(req.FullName) > 100 {
		return fmt.Errorf("姓名长度不能超过100个字符")
	}

	return nil
}

// validateLoginRequest 验证登录请求
func (s *AuthService) validateLoginRequest(req *models.LoginRequest) error {
	if req.Username == "" {
		return fmt.Errorf("用户名/邮箱不能为空")
	}

	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}

	return nil
}

// isEmail 判断是否为邮箱格式
func (s *AuthService) isEmail(str string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(str))
}
