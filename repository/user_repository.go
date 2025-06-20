package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"familytree/models"
)

// UserRepository用户存储库方法 - 扩展SQLiteRepository

// CreateUser 创建用户
func (r *SQLiteRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, password, full_name, avatar, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.FullName,
		user.Avatar,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取新用户ID失败: %v", err)
	}

	user.UserID = int(id)
	return user, nil
}

// GetUserByID 根据ID获取用户
func (r *SQLiteRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT user_id, username, email, password, full_name, avatar, is_active, created_at, updated_at
		FROM users WHERE user_id = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.Avatar,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *SQLiteRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT user_id, username, email, password, full_name, avatar, is_active, created_at, updated_at
		FROM users WHERE username = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.Avatar,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *SQLiteRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT user_id, username, email, password, full_name, avatar, is_active, created_at, updated_at
		FROM users WHERE email = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.Avatar,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func (r *SQLiteRepository) UpdateUser(ctx context.Context, id int, user *models.User) (*models.User, error) {
	query := `
		UPDATE users 
		SET username = ?, email = ?, full_name = ?, avatar = ?, is_active = ?, updated_at = ?
		WHERE user_id = ?
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.FullName,
		user.Avatar,
		user.IsActive,
		user.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("更新用户失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("检查更新结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	user.UserID = id
	return user, nil
}

// UpdatePassword 更新用户密码
func (r *SQLiteRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
		UPDATE users 
		SET password = ?, updated_at = ?
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查更新结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// DeleteUser 删除用户
func (r *SQLiteRepository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查删除结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// FamilyTreeRepository 家族树存储库方法

// CreateFamilyTree 创建家族树
func (r *SQLiteRepository) CreateFamilyTree(ctx context.Context, familyTree *models.UserFamilyTree) (*models.UserFamilyTree, error) {
	// 如果设置为默认家族树，先取消其他家族树的默认状态
	if familyTree.IsDefault {
		updateQuery := `UPDATE user_family_trees SET is_default = 0 WHERE user_id = ?`
		_, err := r.db.ExecContext(ctx, updateQuery, familyTree.UserID)
		if err != nil {
			return nil, fmt.Errorf("更新默认家族树状态失败: %v", err)
		}
	}

	query := `
		INSERT INTO user_family_trees (user_id, family_tree_name, description, root_person_id, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	familyTree.CreatedAt = now
	familyTree.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		familyTree.UserID,
		familyTree.FamilyTreeName,
		familyTree.Description,
		familyTree.RootPersonID,
		familyTree.IsDefault,
		familyTree.CreatedAt,
		familyTree.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("创建家族树失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取新家族树ID失败: %v", err)
	}

	familyTree.FamilyTreeID = int(id)
	return familyTree, nil
}

// GetFamilyTreeByID 根据ID获取家族树
func (r *SQLiteRepository) GetFamilyTreeByID(ctx context.Context, id int) (*models.UserFamilyTree, error) {
	query := `
		SELECT family_tree_id, user_id, family_tree_name, description, root_person_id, is_default, created_at, updated_at
		FROM user_family_trees WHERE family_tree_id = ?
	`

	var familyTree models.UserFamilyTree
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&familyTree.FamilyTreeID,
		&familyTree.UserID,
		&familyTree.FamilyTreeName,
		&familyTree.Description,
		&familyTree.RootPersonID,
		&familyTree.IsDefault,
		&familyTree.CreatedAt,
		&familyTree.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("家族树不存在")
		}
		return nil, fmt.Errorf("查询家族树失败: %v", err)
	}

	return &familyTree, nil
}

// GetUserFamilyTrees 获取用户的所有家族树
func (r *SQLiteRepository) GetUserFamilyTrees(ctx context.Context, userID int) ([]models.UserFamilyTree, error) {
	query := `
		SELECT family_tree_id, user_id, family_tree_name, description, root_person_id, is_default, created_at, updated_at
		FROM user_family_trees WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户家族树失败: %v", err)
	}
	defer rows.Close()

	var familyTrees []models.UserFamilyTree
	for rows.Next() {
		var familyTree models.UserFamilyTree
		err := rows.Scan(
			&familyTree.FamilyTreeID,
			&familyTree.UserID,
			&familyTree.FamilyTreeName,
			&familyTree.Description,
			&familyTree.RootPersonID,
			&familyTree.IsDefault,
			&familyTree.CreatedAt,
			&familyTree.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描家族树数据失败: %v", err)
		}
		familyTrees = append(familyTrees, familyTree)
	}

	return familyTrees, nil
}

// GetDefaultFamilyTree 获取用户的默认家族树
func (r *SQLiteRepository) GetDefaultFamilyTree(ctx context.Context, userID int) (*models.UserFamilyTree, error) {
	query := `
		SELECT family_tree_id, user_id, family_tree_name, description, root_person_id, is_default, created_at, updated_at
		FROM user_family_trees WHERE user_id = ? AND is_default = 1
	`

	var familyTree models.UserFamilyTree
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&familyTree.FamilyTreeID,
		&familyTree.UserID,
		&familyTree.FamilyTreeName,
		&familyTree.Description,
		&familyTree.RootPersonID,
		&familyTree.IsDefault,
		&familyTree.CreatedAt,
		&familyTree.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户没有默认家族树")
		}
		return nil, fmt.Errorf("查询默认家族树失败: %v", err)
	}

	return &familyTree, nil
}

// SetDefaultFamilyTree 设置默认家族树
func (r *SQLiteRepository) SetDefaultFamilyTree(ctx context.Context, userID int, familyTreeID int) error {
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 取消所有家族树的默认状态
	_, err = tx.ExecContext(ctx, `UPDATE user_family_trees SET is_default = 0 WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("清除默认状态失败: %v", err)
	}

	// 设置新的默认家族树
	result, err := tx.ExecContext(ctx, `UPDATE user_family_trees SET is_default = 1 WHERE user_id = ? AND family_tree_id = ?`, userID, familyTreeID)
	if err != nil {
		return fmt.Errorf("设置默认家族树失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查更新结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("家族树不存在或不属于该用户")
	}

	// 提交事务
	return tx.Commit()
}

// UpdateFamilyTree 更新家族树信息
func (r *SQLiteRepository) UpdateFamilyTree(ctx context.Context, id int, familyTree *models.UserFamilyTree) (*models.UserFamilyTree, error) {
	query := `
		UPDATE user_family_trees 
		SET family_tree_name = ?, description = ?, root_person_id = ?, updated_at = ?
		WHERE family_tree_id = ?
	`

	familyTree.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		familyTree.FamilyTreeName,
		familyTree.Description,
		familyTree.RootPersonID,
		familyTree.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("更新家族树失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("检查更新结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("家族树不存在")
	}

	familyTree.FamilyTreeID = id
	return familyTree, nil
}

// DeleteFamilyTree 删除家族树
func (r *SQLiteRepository) DeleteFamilyTree(ctx context.Context, id int) error {
	query := `DELETE FROM user_family_trees WHERE family_tree_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除家族树失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查删除结果失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("家族树不存在")
	}

	return nil
}
