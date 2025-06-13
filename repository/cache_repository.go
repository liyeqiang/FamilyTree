package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"familytree/config"
	"familytree/models"

	"github.com/redis/go-redis/v9"
)

// CacheRepository Redis缓存仓库
type CacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCacheRepository 创建新的缓存仓库
func NewCacheRepository(config *config.RedisConfig, ttl time.Duration) (*CacheRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &CacheRepository{
		client: client,
		ttl:    ttl,
	}, nil
}

// Close 关闭Redis连接
func (r *CacheRepository) Close() error {
	return r.client.Close()
}

// GetIndividual 获取个人缓存
func (r *CacheRepository) GetIndividual(ctx context.Context, id int) (*models.Individual, error) {
	key := fmt.Sprintf("individual:%d", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var individual models.Individual
	if err := json.Unmarshal(data, &individual); err != nil {
		return nil, err
	}

	return &individual, nil
}

// SetIndividual 设置个人缓存
func (r *CacheRepository) SetIndividual(ctx context.Context, individual *models.Individual) error {
	key := fmt.Sprintf("individual:%d", individual.IndividualID)
	data, err := json.Marshal(individual)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// DeleteIndividual 删除个人缓存
func (r *CacheRepository) DeleteIndividual(ctx context.Context, id int) error {
	key := fmt.Sprintf("individual:%d", id)
	return r.client.Del(ctx, key).Err()
}

// GetFamily 获取家庭缓存
func (r *CacheRepository) GetFamily(ctx context.Context, id int) (*models.Family, error) {
	key := fmt.Sprintf("family:%d", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var family models.Family
	if err := json.Unmarshal(data, &family); err != nil {
		return nil, err
	}

	return &family, nil
}

// SetFamily 设置家庭缓存
func (r *CacheRepository) SetFamily(ctx context.Context, family *models.Family) error {
	key := fmt.Sprintf("family:%d", family.FamilyID)
	data, err := json.Marshal(family)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// DeleteFamily 删除家庭缓存
func (r *CacheRepository) DeleteFamily(ctx context.Context, id int) error {
	key := fmt.Sprintf("family:%d", id)
	return r.client.Del(ctx, key).Err()
}

// GetFamilyTree 获取家谱树缓存
func (r *CacheRepository) GetFamilyTree(ctx context.Context, rootID int) (*models.FamilyTreeNode, error) {
	key := fmt.Sprintf("familytree:%d", rootID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var tree models.FamilyTreeNode
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}

	return &tree, nil
}

// SetFamilyTree 设置家谱树缓存
func (r *CacheRepository) SetFamilyTree(ctx context.Context, rootID int, tree *models.FamilyTreeNode) error {
	key := fmt.Sprintf("familytree:%d", rootID)
	data, err := json.Marshal(tree)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// DeleteFamilyTree 删除家谱树缓存
func (r *CacheRepository) DeleteFamilyTree(ctx context.Context, rootID int) error {
	key := fmt.Sprintf("familytree:%d", rootID)
	return r.client.Del(ctx, key).Err()
}

// ClearAll 清除所有缓存
func (r *CacheRepository) ClearAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// BatchGetIndividuals 批量获取个人缓存
func (r *CacheRepository) BatchGetIndividuals(ctx context.Context, ids []int) (map[int]*models.Individual, error) {
	pipe := r.client.Pipeline()
	cmds := make(map[int]*redis.StringCmd)

	// 创建管道命令
	for _, id := range ids {
		key := fmt.Sprintf("individual:%d", id)
		cmds[id] = pipe.Get(ctx, key)
	}

	// 执行管道
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	// 处理结果
	result := make(map[int]*models.Individual)
	for id, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var individual models.Individual
		if err := json.Unmarshal(data, &individual); err != nil {
			return nil, err
		}
		result[id] = &individual
	}

	return result, nil
}

// BatchSetIndividuals 批量设置个人缓存
func (r *CacheRepository) BatchSetIndividuals(ctx context.Context, individuals map[int]*models.Individual) error {
	pipe := r.client.Pipeline()

	for _, individual := range individuals {
		key := fmt.Sprintf("individual:%d", individual.IndividualID)
		data, err := json.Marshal(individual)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, data, r.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BatchDeleteIndividuals 批量删除个人缓存
func (r *CacheRepository) BatchDeleteIndividuals(ctx context.Context, ids []int) error {
	pipe := r.client.Pipeline()

	for _, id := range ids {
		key := fmt.Sprintf("individual:%d", id)
		pipe.Del(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	return err
}
