package handlers

import (
	"encoding/json"
	"net/http"

	"familytree/interfaces"
	"familytree/models"
	"familytree/pkg/middleware"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService interfaces.AuthService
	userService interfaces.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService interfaces.AuthService, userService interfaces.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求数据格式错误",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
			Code:    "REGISTER_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    user,
		Message: "注册成功",
	})
}

// Login 用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求数据格式错误",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	loginResponse, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: err.Error(),
			Code:    "LOGIN_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    loginResponse,
		Message: "登录成功",
	})
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求数据格式错误",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if req.RefreshToken == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "刷新令牌不能为空",
			Code:    "MISSING_REFRESH_TOKEN",
		})
		return
	}

	loginResponse, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: err.Error(),
			Code:    "REFRESH_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    loginResponse,
		Message: "令牌刷新成功",
	})
}

// GetProfile 获取用户资料
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "用户未认证",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userInfo, err := h.userService.GetByID(r.Context(), user.UserID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    userInfo,
	})
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "用户未认证",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req struct {
		FullName string  `json:"full_name"`
		Avatar   *string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求数据格式错误",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// 获取当前用户信息
	currentUser, err := h.userService.GetByID(r.Context(), user.UserID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	// 更新字段
	if req.FullName != "" {
		currentUser.FullName = req.FullName
	}
	if req.Avatar != nil {
		currentUser.Avatar = req.Avatar
	}

	updatedUser, err := h.userService.Update(r.Context(), user.UserID, currentUser)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
			Code:    "UPDATE_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    updatedUser,
		Message: "资料更新成功",
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "用户未认证",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求数据格式错误",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "旧密码和新密码不能为空",
			Code:    "MISSING_PASSWORD",
		})
		return
	}

	if len(req.NewPassword) < 6 {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "新密码长度至少6个字符",
			Code:    "PASSWORD_TOO_SHORT",
		})
		return
	}

	err := h.userService.ChangePassword(r.Context(), user.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
			Code:    "PASSWORD_CHANGE_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "密码修改成功",
	})
}

// Logout 用户登出（客户端处理，清除令牌）
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 在无状态JWT系统中，登出主要由客户端处理
	// 服务端可以实现令牌黑名单，但这里简化处理
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "登出成功",
	})
}

// ValidateToken 验证令牌（用于其他服务验证用户）
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "令牌无效",
			Code:    "INVALID_TOKEN",
		})
		return
	}

	userInfo, err := h.userService.GetByID(r.Context(), user.UserID)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "用户不存在",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"valid": true,
			"user":  userInfo,
		},
		Message: "令牌有效",
	})
}
