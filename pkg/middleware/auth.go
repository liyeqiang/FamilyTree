package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"familytree/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret JWT密钥
var JWTSecret = []byte("your-secret-key-change-in-production")

// ContextKey 上下文键类型
type ContextKey string

const (
	// UserContextKey 用户上下文键
	UserContextKey ContextKey = "user"
	// FamilyTreeContextKey 家族树上下文键
	FamilyTreeContextKey ContextKey = "family_tree"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 跳过认证的路径
		skipPaths := []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/health",
			"/docs",
			"/static/",
			"/ui",
		}

		// 精确匹配的路径（不使用前缀匹配）
		exactSkipPaths := []string{
			"/",
		}

		// 检查是否跳过认证 - 精确匹配
		for _, path := range exactSkipPaths {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 检查是否跳过认证 - 前缀匹配
		for _, path := range skipPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 获取认证头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "未提供认证令牌", http.StatusUnauthorized)
			return
		}

		// 检查Bearer令牌格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "认证令牌格式无效", http.StatusUnauthorized)
			return
		}

		// 验证JWT令牌
		claims, err := ValidateJWTToken(tokenString)
		if err != nil {
			http.Error(w, "无效的认证令牌", http.StatusUnauthorized)
			return
		}

		// 将用户信息添加到上下文
		ctx := context.WithValue(r.Context(), UserContextKey, &models.AuthContext{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateJWTToken 生成JWT令牌
func GenerateJWTToken(user *models.User) (string, string, error) {
	// 访问令牌，有效期24小时
	accessClaims := &models.JWTClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   fmt.Sprintf("%d", user.UserID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(JWTSecret)
	if err != nil {
		return "", "", err
	}

	// 刷新令牌，有效期7天
	refreshClaims := &models.JWTClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   fmt.Sprintf("%d", user.UserID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(JWTSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateJWTToken 验证JWT令牌
func ValidateJWTToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) (*models.AuthContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.AuthContext)
	return user, ok
}

// RequireAuth 要求认证的中间件函数
func RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "用户未认证", http.StatusUnauthorized)
			return
		}

		// 可以在这里添加额外的权限检查
		_ = user // 使用用户信息

		handler(w, r)
	}
}

// FamilyTreeAccessMiddleware 家族树访问权限中间件
func FamilyTreeAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "用户未认证", http.StatusUnauthorized)
			return
		}

		// 这里可以添加家族树访问权限检查
		// 例如检查用户是否有权访问特定的家族树

		// 将家族树信息添加到上下文（如果需要）
		ctx := context.WithValue(r.Context(), FamilyTreeContextKey, map[string]interface{}{
			"user_id": user.UserID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
