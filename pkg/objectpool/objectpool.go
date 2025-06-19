package objectpool

import (
	"familytree/models"
	"sync"
	"time"
)

// ObjectPool 通用对象池
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool 创建新的对象池
func NewObjectPool(new func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: new,
		},
	}
}

// Get 从对象池获取对象
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put 将对象放回对象池
func (p *ObjectPool) Put(x interface{}) {
	p.pool.Put(x)
}

// IndividualPool 个人信息对象池
type IndividualPool struct {
	*ObjectPool
}

// NewIndividualPool 创建个人信息对象池
func NewIndividualPool() *IndividualPool {
	return &IndividualPool{
		ObjectPool: NewObjectPool(func() interface{} {
			return &models.Individual{}
		}),
	}
}

// Get 从对象池获取个人信息对象
func (p *IndividualPool) Get() *models.Individual {
	return p.ObjectPool.Get().(*models.Individual)
}

// Put 将个人信息对象放回对象池
func (p *IndividualPool) Put(x *models.Individual) {
	// 清空对象
	x.IndividualID = 0
	x.FullName = ""
	x.Gender = ""
	x.BirthDate = nil
	x.BirthPlace = nil
	x.BirthPlaceID = nil
	x.DeathDate = nil
	x.DeathPlace = nil
	x.DeathPlaceID = nil
	x.Occupation = ""
	x.Notes = ""
	x.PhotoURL = nil
	x.FatherID = nil
	x.MotherID = nil
	x.CreatedAt = time.Time{}
	x.UpdatedAt = time.Time{}
	x.BirthPlaceObj = nil
	x.DeathPlaceObj = nil
	x.Father = nil
	x.Mother = nil
	x.Children = nil
	x.MarriageOrder = 0

	p.ObjectPool.Put(x)
}

// FamilyPool 家庭对象池
type FamilyPool struct {
	*ObjectPool
}

// NewFamilyPool 创建家庭对象池
func NewFamilyPool() *FamilyPool {
	return &FamilyPool{
		ObjectPool: NewObjectPool(func() interface{} {
			return &models.Family{}
		}),
	}
}

// Get 从对象池获取家庭对象
func (p *FamilyPool) Get() *models.Family {
	return p.ObjectPool.Get().(*models.Family)
}

// Put 将家庭对象放回对象池
func (p *FamilyPool) Put(x *models.Family) {
	// 清空对象
	x.FamilyID = 0
	x.HusbandID = nil
	x.WifeID = nil
	x.MarriageOrder = 0
	x.MarriageDate = nil
	x.MarriagePlaceID = nil
	x.DivorceDate = nil
	x.Notes = ""
	x.CreatedAt = time.Time{}
	x.UpdatedAt = time.Time{}
	x.Husband = nil
	x.Wife = nil
	x.MarriagePlace = nil
	x.Children = nil

	p.ObjectPool.Put(x)
}

// FamilyTreeNodePool 家族树节点对象池
type FamilyTreeNodePool struct {
	*ObjectPool
}

// NewFamilyTreeNodePool 创建家族树节点对象池
func NewFamilyTreeNodePool() *FamilyTreeNodePool {
	return &FamilyTreeNodePool{
		ObjectPool: NewObjectPool(func() interface{} {
			return &models.FamilyTreeNode{}
		}),
	}
}

// Get 从对象池获取家族树节点对象
func (p *FamilyTreeNodePool) Get() *models.FamilyTreeNode {
	return p.ObjectPool.Get().(*models.FamilyTreeNode)
}

// Put 将家族树节点对象放回对象池
func (p *FamilyTreeNodePool) Put(x *models.FamilyTreeNode) {
	// 清空对象
	x.Individual = nil
	x.Spouse = nil
	x.Children = nil
	x.Parents = nil

	p.ObjectPool.Put(x)
}
