package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"familytree/models"
)

// SQLiteRepository SQLite数据库存储库实现
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository 创建新的SQLite存储库
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Individual相关方法

// CreateIndividual 创建个人信息
func (r *SQLiteRepository) CreateIndividual(ctx context.Context, individual *models.Individual) (*models.Individual, error) {
	query := `
		INSERT INTO individuals (full_name, gender, birth_date, birth_place_id, death_date, 
		death_place_id, occupation, notes, photo_url, father_id, mother_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		individual.FullName, individual.Gender, individual.BirthDate,
		individual.BirthPlaceID, individual.DeathDate, individual.DeathPlaceID,
		individual.Occupation, individual.Notes, individual.PhotoURL,
		individual.FatherID, individual.MotherID)
	
	if err != nil {
		return nil, fmt.Errorf("创建个人信息失败: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取新插入ID失败: %v", err)
	}
	
	individual.IndividualID = int(id)
	individual.CreatedAt = time.Now()
	individual.UpdatedAt = time.Now()
	
	return individual, nil
}

// GetIndividualByID 根据ID获取个人信息
func (r *SQLiteRepository) GetIndividualByID(ctx context.Context, id int) (*models.Individual, error) {
	query := `
		SELECT individual_id, full_name, gender, birth_date, birth_place_id, death_date,
		death_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
		FROM individuals WHERE individual_id = ?
	`
	
	var individual models.Individual
	row := r.db.QueryRowContext(ctx, query, id)
	
	err := row.Scan(
		&individual.IndividualID, &individual.FullName, &individual.Gender,
		&individual.BirthDate, &individual.BirthPlaceID, &individual.DeathDate,
		&individual.DeathPlaceID, &individual.Occupation, &individual.Notes,
		&individual.PhotoURL, &individual.FatherID, &individual.MotherID,
		&individual.CreatedAt, &individual.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("个人信息不存在")
		}
		return nil, fmt.Errorf("查询个人信息失败: %v", err)
	}
	
	return &individual, nil
}

// UpdateIndividual 更新个人信息
func (r *SQLiteRepository) UpdateIndividual(ctx context.Context, id int, individual *models.Individual) (*models.Individual, error) {
	query := `
		UPDATE individuals SET 
		full_name = ?, gender = ?, birth_date = ?, birth_place_id = ?, death_date = ?,
		death_place_id = ?, occupation = ?, notes = ?, photo_url = ?, father_id = ?, mother_id = ?
		WHERE individual_id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query,
		individual.FullName, individual.Gender, individual.BirthDate,
		individual.BirthPlaceID, individual.DeathDate, individual.DeathPlaceID,
		individual.Occupation, individual.Notes, individual.PhotoURL,
		individual.FatherID, individual.MotherID, id)
	
	if err != nil {
		return nil, fmt.Errorf("更新个人信息失败: %v", err)
	}
	
	return r.GetIndividualByID(ctx, id)
}

// DeleteIndividual 删除个人信息
func (r *SQLiteRepository) DeleteIndividual(ctx context.Context, id int) error {
	query := `DELETE FROM individuals WHERE individual_id = ?`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除个人信息失败: %v", err)
	}
	
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	
	if affected == 0 {
		return fmt.Errorf("个人信息不存在")
	}
	
	return nil
}

// SearchIndividuals 搜索个人信息
func (r *SQLiteRepository) SearchIndividuals(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	var individuals []models.Individual
	var args []interface{}
	
	baseQuery := `
		SELECT individual_id, full_name, gender, birth_date, birth_place_id, death_date,
		death_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
		FROM individuals
	`
	
	countQuery := "SELECT COUNT(*) FROM individuals"
	
	if query != "" {
		whereClause := " WHERE full_name LIKE ? OR occupation LIKE ? OR notes LIKE ?"
		baseQuery += whereClause
		countQuery += whereClause
		searchTerm := "%" + query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}
	
	// 获取总数
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("获取搜索结果总数失败: %v", err)
	}
	
	// 获取分页结果
	baseQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索个人信息失败: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var individual models.Individual
		err := rows.Scan(
			&individual.IndividualID, &individual.FullName, &individual.Gender,
			&individual.BirthDate, &individual.BirthPlaceID, &individual.DeathDate,
			&individual.DeathPlaceID, &individual.Occupation, &individual.Notes,
			&individual.PhotoURL, &individual.FatherID, &individual.MotherID,
			&individual.CreatedAt, &individual.UpdatedAt)
		
		if err != nil {
			return nil, 0, fmt.Errorf("扫描搜索结果失败: %v", err)
		}
		
		individuals = append(individuals, individual)
	}
	
	return individuals, total, nil
}

// GetIndividualsByParentID 根据父母ID获取子女
func (r *SQLiteRepository) GetIndividualsByParentID(ctx context.Context, parentID int) ([]models.Individual, error) {
	query := `
		SELECT individual_id, full_name, gender, birth_date, birth_place_id, death_date,
		death_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
		FROM individuals WHERE father_id = ? OR mother_id = ?
		ORDER BY birth_date
	`
	
	rows, err := r.db.QueryContext(ctx, query, parentID, parentID)
	if err != nil {
		return nil, fmt.Errorf("查询子女信息失败: %v", err)
	}
	defer rows.Close()
	
	var individuals []models.Individual
	for rows.Next() {
		var individual models.Individual
		err := rows.Scan(
			&individual.IndividualID, &individual.FullName, &individual.Gender,
			&individual.BirthDate, &individual.BirthPlaceID, &individual.DeathDate,
			&individual.DeathPlaceID, &individual.Occupation, &individual.Notes,
			&individual.PhotoURL, &individual.FatherID, &individual.MotherID,
			&individual.CreatedAt, &individual.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("扫描子女信息失败: %v", err)
		}
		
		individuals = append(individuals, individual)
	}
	
	return individuals, nil
}

// GetIndividualsByIDs 根据ID列表获取个人信息
func (r *SQLiteRepository) GetIndividualsByIDs(ctx context.Context, ids []int) ([]models.Individual, error) {
	if len(ids) == 0 {
		return []models.Individual{}, nil
	}
	
	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf(`
		SELECT individual_id, full_name, gender, birth_date, birth_place_id, death_date,
		death_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
		FROM individuals WHERE individual_id IN (%s)
	`, placeholders)
	
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询个人信息列表失败: %v", err)
	}
	defer rows.Close()
	
	var individuals []models.Individual
	for rows.Next() {
		var individual models.Individual
		err := rows.Scan(
			&individual.IndividualID, &individual.FullName, &individual.Gender,
			&individual.BirthDate, &individual.BirthPlaceID, &individual.DeathDate,
			&individual.DeathPlaceID, &individual.Occupation, &individual.Notes,
			&individual.PhotoURL, &individual.FatherID, &individual.MotherID,
			&individual.CreatedAt, &individual.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("扫描个人信息失败: %v", err)
		}
		
		individuals = append(individuals, individual)
	}
	
	return individuals, nil
}

// GetParents 获取个人的父母信息
func (r *SQLiteRepository) GetParents(ctx context.Context, individualID int) (*models.Individual, *models.Individual, error) {
	individual, err := r.GetIndividualByID(ctx, individualID)
	if err != nil {
		return nil, nil, err
	}
	
	var father, mother *models.Individual
	
	if individual.FatherID != nil {
		father, _ = r.GetIndividualByID(ctx, *individual.FatherID)
	}
	
	if individual.MotherID != nil {
		mother, _ = r.GetIndividualByID(ctx, *individual.MotherID)
	}
	
	return father, mother, nil
}

// GetSiblings 获取兄弟姐妹
func (r *SQLiteRepository) GetSiblings(ctx context.Context, individualID int) ([]models.Individual, error) {
	individual, err := r.GetIndividualByID(ctx, individualID)
	if err != nil {
		return nil, err
	}
	
	query := `
		SELECT individual_id, full_name, gender, birth_date, birth_place_id, death_date,
		death_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
		FROM individuals 
		WHERE individual_id != ? AND (
			(father_id = ? AND father_id IS NOT NULL) OR 
			(mother_id = ? AND mother_id IS NOT NULL)
		)
		ORDER BY birth_date
	`
	
	rows, err := r.db.QueryContext(ctx, query, individualID, individual.FatherID, individual.MotherID)
	if err != nil {
		return nil, fmt.Errorf("查询兄弟姐妹失败: %v", err)
	}
	defer rows.Close()
	
	var siblings []models.Individual
	for rows.Next() {
		var sibling models.Individual
		err := rows.Scan(
			&sibling.IndividualID, &sibling.FullName, &sibling.Gender,
			&sibling.BirthDate, &sibling.BirthPlaceID, &sibling.DeathDate,
			&sibling.DeathPlaceID, &sibling.Occupation, &sibling.Notes,
			&sibling.PhotoURL, &sibling.FatherID, &sibling.MotherID,
			&sibling.CreatedAt, &sibling.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("扫描兄弟姐妹信息失败: %v", err)
		}
		
		siblings = append(siblings, sibling)
	}
	
	return siblings, nil
}

// GetSpouses 获取配偶
func (r *SQLiteRepository) GetSpouses(ctx context.Context, individualID int) ([]models.Individual, error) {
	query := `
		SELECT CASE 
			WHEN f.husband_id = ? THEN f.wife_id 
			ELSE f.husband_id 
		END as spouse_id
		FROM families f 
		WHERE f.husband_id = ? OR f.wife_id = ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, individualID, individualID, individualID)
	if err != nil {
		return nil, fmt.Errorf("查询配偶失败: %v", err)
	}
	defer rows.Close()
	
	var spouseIDs []int
	for rows.Next() {
		var spouseID *int
		err := rows.Scan(&spouseID)
		if err != nil {
			return nil, fmt.Errorf("扫描配偶ID失败: %v", err)
		}
		if spouseID != nil {
			spouseIDs = append(spouseIDs, *spouseID)
		}
	}
	
	if len(spouseIDs) == 0 {
		return []models.Individual{}, nil
	}
	
	return r.GetIndividualsByIDs(ctx, spouseIDs)
}

// GetAncestors 获取祖先
func (r *SQLiteRepository) GetAncestors(ctx context.Context, individualID int, generations int) ([]models.Individual, error) {
	var ancestors []models.Individual
	visited := make(map[int]bool)
	
	var getAncestorsRecursive func(int, int) error
	getAncestorsRecursive = func(id int, gen int) error {
		if gen <= 0 || visited[id] {
			return nil
		}
		
		visited[id] = true
		father, mother, err := r.GetParents(ctx, id)
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
	
	err := getAncestorsRecursive(individualID, generations)
	return ancestors, err
}

// GetDescendants 获取后代
func (r *SQLiteRepository) GetDescendants(ctx context.Context, individualID int, generations int) ([]models.Individual, error) {
	var descendants []models.Individual
	visited := make(map[int]bool)
	
	var getDescendantsRecursive func(int, int) error
	getDescendantsRecursive = func(id int, gen int) error {
		if gen <= 0 || visited[id] {
			return nil
		}
		
		visited[id] = true
		children, err := r.GetIndividualsByParentID(ctx, id)
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
	
	err := getDescendantsRecursive(individualID, generations)
	return descendants, err
}

// BuildFamilyTree 构建家族树
func (r *SQLiteRepository) BuildFamilyTree(ctx context.Context, rootID int, generations int) (*models.FamilyTreeNode, error) {
	individual, err := r.GetIndividualByID(ctx, rootID)
	if err != nil {
		return nil, err
	}
	
	node := &models.FamilyTreeNode{
		Individual: individual,
	}
	
	if generations > 0 {
		children, err := r.GetIndividualsByParentID(ctx, rootID)
		if err != nil {
			return nil, err
		}
		
		for _, child := range children {
			childNode, err := r.BuildFamilyTree(ctx, child.IndividualID, generations-1)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, *childNode)
		}
	}
	
	return node, nil
} 