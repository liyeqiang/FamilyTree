package di

import (
	"errors"
	"reflect"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	services map[reflect.Type]interface{}
	mu       sync.RWMutex
}

// NewContainer 创建新的依赖注入容器
func NewContainer() *Container {
	return &Container{
		services: make(map[reflect.Type]interface{}),
	}
}

// Register 注册服务
func (c *Container) Register(service interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(service)
	if t.Kind() != reflect.Ptr {
		return errors.New("service must be a pointer")
	}

	c.services[t] = service
	return nil
}

// Get 获取服务
func (c *Container) Get(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if service, ok := c.services[serviceType]; ok {
		return service, nil
	}

	return nil, errors.New("service not found")
}

// Resolve 解析依赖
func (c *Container) Resolve(serviceType reflect.Type) (interface{}, error) {
	// 检查服务是否已注册
	if service, err := c.Get(serviceType); err == nil {
		return service, nil
	}

	// 创建新实例
	instance := reflect.New(serviceType.Elem()).Interface()

	// 获取所有字段
	value := reflect.ValueOf(instance).Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := field.Type()

		// 检查字段是否可注入
		if field.CanSet() && fieldType.Kind() == reflect.Ptr {
			// 递归解析依赖
			if dependency, err := c.Resolve(fieldType); err == nil {
				field.Set(reflect.ValueOf(dependency))
			}
		}
	}

	// 注册新实例
	if err := c.Register(instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// Provider 服务提供者接口
type Provider interface {
	Register(container *Container) error
}

// ServiceProvider 服务提供者
type ServiceProvider struct {
	container *Container
}

// NewServiceProvider 创建新的服务提供者
func NewServiceProvider() *ServiceProvider {
	return &ServiceProvider{
		container: NewContainer(),
	}
}

// Register 注册服务提供者
func (p *ServiceProvider) Register(provider Provider) error {
	return provider.Register(p.container)
}

// Get 获取服务
func (p *ServiceProvider) Get(serviceType reflect.Type) (interface{}, error) {
	return p.container.Get(serviceType)
}

// Resolve 解析依赖
func (p *ServiceProvider) Resolve(serviceType reflect.Type) (interface{}, error) {
	return p.container.Resolve(serviceType)
}

// Singleton 单例服务
type Singleton struct {
	instance interface{}
	mu       sync.Mutex
}

// NewSingleton 创建新的单例服务
func NewSingleton() *Singleton {
	return &Singleton{}
}

// Get 获取单例实例
func (s *Singleton) Get(create func() interface{}) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.instance == nil {
		s.instance = create()
	}

	return s.instance
}

// Factory 工厂函数
type Factory func(container *Container) (interface{}, error)

// ServiceFactory 服务工厂
type ServiceFactory struct {
	factory Factory
}

// NewServiceFactory 创建新的服务工厂
func NewServiceFactory(factory Factory) *ServiceFactory {
	return &ServiceFactory{
		factory: factory,
	}
}

// Create 创建服务实例
func (f *ServiceFactory) Create(container *Container) (interface{}, error) {
	return f.factory(container)
}
