package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"familytree/models"

	_ "modernc.org/sqlite"
)

// SQLiteRepository SQLite数据库存储库实现
type SQLiteRepository struct {
	db *sql.DB

	// 预处理语句
	stmtCache struct {
		sync.RWMutex
		statements map[string]*sql.Stmt
	}

	// 连接池配置
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

// NewSQLiteRepository 创建新的SQLite存储库
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	repo := &SQLiteRepository{
		db: db,
		stmtCache: struct {
			sync.RWMutex
			statements map[string]*sql.Stmt
		}{
			statements: make(map[string]*sql.Stmt),
		},
		maxOpenConns:    25,
		maxIdleConns:    10,
		connMaxLifetime: time.Hour,
	}

	// 配置连接池
	db.SetMaxOpenConns(repo.maxOpenConns)
	db.SetMaxIdleConns(repo.maxIdleConns)
	db.SetConnMaxLifetime(repo.connMaxLifetime)

	// 初始化预处理语句
	if err := repo.initPreparedStatements(); err != nil {
		return nil, fmt.Errorf("初始化预处理语句失败: %v", err)
	}

	return repo, nil
}

// initPreparedStatements 初始化所有预处理语句
func (r *SQLiteRepository) initPreparedStatements() error {
	statements := map[string]string{
		"get_individual_by_id": `
			SELECT individual_id, full_name, gender, birth_date, birth_place, birth_place_id,
			       death_date, death_place, death_place_id, burial_place_id,
			       occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
			FROM individuals WHERE individual_id = ?
		`,
		"search_individuals": `
			SELECT individual_id, full_name, gender, birth_date, birth_place, birth_place_id,
			       death_date, death_place, death_place_id, burial_place_id,
			       occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
			FROM individuals 
			WHERE full_name LIKE ? OR notes LIKE ?
			LIMIT ? OFFSET ?
		`,
		"create_individual": `
			INSERT INTO individuals (
				full_name, gender, birth_date, birth_place, birth_place_id,
				death_date, death_place, death_place_id, burial_place_id,
				occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		"update_individual": `
			UPDATE individuals SET
				full_name = ?, gender = ?, birth_date = ?, birth_place = ?,
				birth_place_id = ?, death_date = ?, death_place = ?,
				death_place_id = ?, burial_place_id = ?, occupation = ?, notes = ?,
				photo_url = ?, father_id = ?, mother_id = ?, updated_at = ?
			WHERE individual_id = ?
		`,
		"delete_individual": `
			DELETE FROM individuals WHERE individual_id = ?
		`,
		"get_children_by_parent": `
			SELECT individual_id, full_name, gender, birth_date, birth_place, birth_place_id,
			       death_date, death_place, death_place_id, burial_place_id,
			       occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
			FROM individuals 
			WHERE father_id = ? OR mother_id = ?
		`,
		"get_spouses": `
			SELECT i.individual_id, i.full_name, i.gender, i.birth_date, i.birth_place, i.birth_place_id,
			       i.death_date, i.death_place, i.death_place_id, i.burial_place_id,
			       i.occupation, i.notes, i.photo_url, i.father_id, i.mother_id, i.created_at, i.updated_at
			FROM individuals i
			JOIN families f ON (f.husband_id = i.individual_id OR f.wife_id = i.individual_id)
			WHERE (f.husband_id = ? OR f.wife_id = ?)
			AND i.individual_id != ?
		`,
	}

	r.stmtCache.Lock()
	defer r.stmtCache.Unlock()

	for name, query := range statements {
		stmt, err := r.db.Prepare(query)
		if err != nil {
			return fmt.Errorf("准备语句 %s 失败: %v", name, err)
		}
		r.stmtCache.statements[name] = stmt
	}

	return nil
}

// getStmt 获取预处理语句
func (r *SQLiteRepository) getStmt(name string) (*sql.Stmt, error) {
	r.stmtCache.RLock()
	stmt, exists := r.stmtCache.statements[name]
	r.stmtCache.RUnlock()

	if !exists {
		return nil, fmt.Errorf("预处理语句 %s 不存在", name)
	}

	return stmt, nil
}

// Individual相关方法

// CreateIndividual 创建个人信息
func (r *SQLiteRepository) CreateIndividual(ctx context.Context, individual *models.Individual) (*models.Individual, error) {
	stmt, err := r.getStmt("create_individual")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	individual.CreatedAt = now
	individual.UpdatedAt = now

	result, err := stmt.ExecContext(ctx,
		individual.FullName,
		individual.Gender,
		individual.BirthDate,
		individual.BirthPlace,
		individual.BirthPlaceID,
		individual.DeathDate,
		individual.DeathPlace,
		individual.DeathPlaceID,
		individual.BurialPlaceID,
		individual.Occupation,
		individual.Notes,
		individual.PhotoURL,
		individual.FatherID,
		individual.MotherID,
		individual.CreatedAt,
		individual.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	individual.IndividualID = int(id)
	return individual, nil
}

// GetIndividualByID 根据ID获取个人信息
func (r *SQLiteRepository) GetIndividualByID(ctx context.Context, id int) (*models.Individual, error) {
	stmt, err := r.getStmt("get_individual_by_id")
	if err != nil {
		return nil, err
	}

	var individual models.Individual
	err = stmt.QueryRowContext(ctx, id).Scan(
		&individual.IndividualID,
		&individual.FullName,
		&individual.Gender,
		&individual.BirthDate,
		&individual.BirthPlace,
		&individual.BirthPlaceID,
		&individual.DeathDate,
		&individual.DeathPlace,
		&individual.DeathPlaceID,
		&individual.BurialPlaceID,
		&individual.Occupation,
		&individual.Notes,
		&individual.PhotoURL,
		&individual.FatherID,
		&individual.MotherID,
		&individual.CreatedAt,
		&individual.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("个人信息不存在")
		}
		return nil, err
	}

	return &individual, nil
}

// UpdateIndividual 更新个人信息
func (r *SQLiteRepository) UpdateIndividual(ctx context.Context, id int, individual *models.Individual) (*models.Individual, error) {
	stmt, err := r.getStmt("update_individual")
	if err != nil {
		return nil, err
	}

	individual.UpdatedAt = time.Now()

	result, err := stmt.ExecContext(ctx,
		individual.FullName,
		individual.Gender,
		individual.BirthDate,
		individual.BirthPlace,
		individual.BirthPlaceID,
		individual.DeathDate,
		individual.DeathPlace,
		individual.DeathPlaceID,
		individual.BurialPlaceID,
		individual.Occupation,
		individual.Notes,
		individual.PhotoURL,
		individual.FatherID,
		individual.MotherID,
		individual.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("个人信息不存在")
	}

	individual.IndividualID = id
	return individual, nil
}

// DeleteIndividual 删除个人信息
func (r *SQLiteRepository) DeleteIndividual(ctx context.Context, id int) error {
	stmt, err := r.getStmt("delete_individual")
	if err != nil {
		return err
	}

	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("个人信息不存在")
	}

	return nil
}

// SearchIndividuals 搜索个人信息
func (r *SQLiteRepository) SearchIndividuals(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	stmt, err := r.getStmt("search_individuals")
	if err != nil {
		return nil, 0, err
	}

	searchPattern := "%" + query + "%"
	rows, err := stmt.QueryContext(ctx, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var individuals []models.Individual
	for rows.Next() {
		var individual models.Individual
		err := rows.Scan(
			&individual.IndividualID,
			&individual.FullName,
			&individual.Gender,
			&individual.BirthDate,
			&individual.BirthPlace,
			&individual.BirthPlaceID,
			&individual.DeathDate,
			&individual.DeathPlace,
			&individual.DeathPlaceID,
			&individual.BurialPlaceID,
			&individual.Occupation,
			&individual.Notes,
			&individual.PhotoURL,
			&individual.FatherID,
			&individual.MotherID,
			&individual.CreatedAt,
			&individual.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		individuals = append(individuals, individual)
	}

	// 获取总数
	var total int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM individuals WHERE full_name LIKE ? OR notes LIKE ?", searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return individuals, total, nil
}

// GetIndividualsByParentID 根据父母ID获取子女
func (r *SQLiteRepository) GetIndividualsByParentID(ctx context.Context, parentID int) ([]models.Individual, error) {
	stmt, err := r.getStmt("get_children_by_parent")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, parentID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var individuals []models.Individual
	for rows.Next() {
		var individual models.Individual
		err := rows.Scan(
			&individual.IndividualID,
			&individual.FullName,
			&individual.Gender,
			&individual.BirthDate,
			&individual.BirthPlace,
			&individual.BirthPlaceID,
			&individual.DeathDate,
			&individual.DeathPlace,
			&individual.DeathPlaceID,
			&individual.BurialPlaceID,
			&individual.Occupation,
			&individual.Notes,
			&individual.PhotoURL,
			&individual.FatherID,
			&individual.MotherID,
			&individual.CreatedAt,
			&individual.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
		SELECT individual_id, full_name, gender, birth_date, birth_place, birth_place_id, death_date,
		death_place, death_place_id, burial_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
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
			&individual.BirthDate, &individual.BirthPlace, &individual.BirthPlaceID, &individual.DeathDate,
			&individual.DeathPlace, &individual.DeathPlaceID, &individual.BurialPlaceID, &individual.Occupation, &individual.Notes,
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
		SELECT individual_id, full_name, gender, birth_date, birth_place, birth_place_id, death_date,
		death_place, death_place_id, burial_place_id, occupation, notes, photo_url, father_id, mother_id, created_at, updated_at
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
			&sibling.BirthDate, &sibling.BirthPlace, &sibling.BirthPlaceID, &sibling.DeathDate,
			&sibling.DeathPlace, &sibling.DeathPlaceID, &sibling.BurialPlaceID, &sibling.Occupation, &sibling.Notes,
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
	stmt, err := r.getStmt("get_spouses")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, individualID, individualID, individualID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spouses []models.Individual
	for rows.Next() {
		var spouse models.Individual
		err := rows.Scan(
			&spouse.IndividualID,
			&spouse.FullName,
			&spouse.Gender,
			&spouse.BirthDate,
			&spouse.BirthPlace,
			&spouse.BirthPlaceID,
			&spouse.DeathDate,
			&spouse.DeathPlace,
			&spouse.DeathPlaceID,
			&spouse.BurialPlaceID,
			&spouse.Occupation,
			&spouse.Notes,
			&spouse.PhotoURL,
			&spouse.FatherID,
			&spouse.MotherID,
			&spouse.CreatedAt,
			&spouse.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		spouses = append(spouses, spouse)
	}

	return spouses, nil
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

// Family相关方法

// CreateFamily 创建家庭关系
func (r *SQLiteRepository) CreateFamily(ctx context.Context, family *models.Family) (*models.Family, error) {
	query := `
		INSERT INTO families (husband_id, wife_id, marriage_order, marriage_date, marriage_place_id, divorce_date, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		family.HusbandID, family.WifeID, family.MarriageOrder, family.MarriageDate,
		family.MarriagePlaceID, family.DivorceDate, family.Notes)

	if err != nil {
		return nil, fmt.Errorf("创建家庭关系失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取新插入ID失败: %v", err)
	}

	family.FamilyID = int(id)
	family.CreatedAt = time.Now()
	family.UpdatedAt = time.Now()

	return family, nil
}

// GetFamilyByID 根据ID获取家庭关系
func (r *SQLiteRepository) GetFamilyByID(ctx context.Context, id int) (*models.Family, error) {
	query := `
		SELECT family_id, husband_id, wife_id, marriage_order, marriage_date, marriage_place_id, 
		divorce_date, notes, created_at, updated_at
		FROM families WHERE family_id = ?
	`

	var family models.Family
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&family.FamilyID, &family.HusbandID, &family.WifeID, &family.MarriageOrder,
		&family.MarriageDate, &family.MarriagePlaceID, &family.DivorceDate,
		&family.Notes, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("家庭关系不存在")
		}
		return nil, fmt.Errorf("查询家庭关系失败: %v", err)
	}

	return &family, nil
}

// UpdateFamily 更新家庭关系
func (r *SQLiteRepository) UpdateFamily(ctx context.Context, id int, family *models.Family) (*models.Family, error) {
	query := `
		UPDATE families SET 
		husband_id = ?, wife_id = ?, marriage_order = ?, marriage_date = ?, marriage_place_id = ?, 
		divorce_date = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE family_id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		family.HusbandID, family.WifeID, family.MarriageOrder, family.MarriageDate,
		family.MarriagePlaceID, family.DivorceDate, family.Notes, id)

	if err != nil {
		return nil, fmt.Errorf("更新家庭关系失败: %v", err)
	}

	return r.GetFamilyByID(ctx, id)
}

// DeleteFamily 删除家庭关系
func (r *SQLiteRepository) DeleteFamily(ctx context.Context, id int) error {
	query := `DELETE FROM families WHERE family_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除家庭关系失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}

	if affected == 0 {
		return fmt.Errorf("家庭关系不存在")
	}

	return nil
}

// GetFamiliesByIndividualID 获取某人参与的所有家庭关系
func (r *SQLiteRepository) GetFamiliesByIndividualID(ctx context.Context, individualID int) ([]models.Family, error) {
	query := `
		SELECT family_id, husband_id, wife_id, marriage_order, marriage_date, marriage_place_id, 
		divorce_date, notes, created_at, updated_at
		FROM families WHERE husband_id = ? OR wife_id = ?
		ORDER BY marriage_order, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, individualID, individualID)
	if err != nil {
		return nil, fmt.Errorf("查询家庭关系失败: %v", err)
	}
	defer rows.Close()

	var families []models.Family
	for rows.Next() {
		var family models.Family
		err := rows.Scan(
			&family.FamilyID, &family.HusbandID, &family.WifeID, &family.MarriageOrder,
			&family.MarriageDate, &family.MarriagePlaceID, &family.DivorceDate,
			&family.Notes, &family.CreatedAt, &family.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("扫描家庭关系失败: %v", err)
		}

		families = append(families, family)
	}

	return families, nil
}

// CreateChild 创建子女关系
func (r *SQLiteRepository) CreateChild(ctx context.Context, child *models.Child) (*models.Child, error) {
	query := `
		INSERT INTO children (family_id, individual_id, relationship_to_parents)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		child.FamilyID, child.IndividualID, child.RelationshipToParents)

	if err != nil {
		return nil, fmt.Errorf("创建子女关系失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取新插入ID失败: %v", err)
	}

	child.ChildID = int(id)
	child.CreatedAt = time.Now()
	child.UpdatedAt = time.Now()

	return child, nil
}

// DeleteChild 删除子女关系
func (r *SQLiteRepository) DeleteChild(ctx context.Context, familyID, individualID int) error {
	query := `DELETE FROM children WHERE family_id = ? AND individual_id = ?`

	result, err := r.db.ExecContext(ctx, query, familyID, individualID)
	if err != nil {
		return fmt.Errorf("删除子女关系失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}

	if affected == 0 {
		return fmt.Errorf("子女关系不存在")
	}

	return nil
}

// GetChildrenByFamilyID 获取家庭的所有子女
func (r *SQLiteRepository) GetChildrenByFamilyID(ctx context.Context, familyID int) ([]models.Child, error) {
	query := `
		SELECT child_id, family_id, individual_id, relationship_to_parents, created_at, updated_at
		FROM children WHERE family_id = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, fmt.Errorf("查询子女关系失败: %v", err)
	}
	defer rows.Close()

	var children []models.Child
	for rows.Next() {
		var child models.Child
		err := rows.Scan(
			&child.ChildID, &child.FamilyID, &child.IndividualID,
			&child.RelationshipToParents, &child.CreatedAt, &child.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("扫描子女关系失败: %v", err)
		}

		children = append(children, child)
	}

	return children, nil
}

// Close 关闭数据库连接
func (r *SQLiteRepository) Close() error {
	r.stmtCache.Lock()
	for _, stmt := range r.stmtCache.statements {
		stmt.Close()
	}
	r.stmtCache.Unlock()
	return r.db.Close()
}
