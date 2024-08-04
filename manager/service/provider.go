package service

// Proxy provides operations on proxy.
import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/repository"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type IProxyProviderService interface {
	GetProvider(context.Context, string) (*model.ProxyProvider, error)
	AddProvider(context.Context, string) (*string, error)
	ListProvider(context.Context, int64, int64) ([]model.ProxyProvider, error)
	UpdateProvider(context.Context, model.ProxyProvider) error
	DeleteProvider(context.Context, string) error
}

type ProxyProviderService struct {
	logger *logrus.Logger
	db     *gorm.DB
}

func NewProxyProviderService(logger *logrus.Logger, db *gorm.DB) *ProxyProviderService {
	return &ProxyProviderService{
		logger: logger,
		db:     db,
	}
}

func (p ProxyProviderService) GetProvider(ctx context.Context, id string) (*model.ProxyProvider, error) {
	var provider repository.ProxyProvider
	result := p.db.Where(repository.ProxyProvider{ProviderId: id}).First(&provider)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("provider %s not exists", id))
	}
	ret_provider := model.ProxyProvider{}
	ret_provider.Id = provider.ProviderId
	ret_provider.Name = provider.Name
	return &ret_provider, nil
}
func (p ProxyProviderService) AddProvider(ctx context.Context, name string) (*string, error) {
	md5_gen := md5.New()
	_, err := md5_gen.Write([]byte(name))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	id := fmt.Sprintf("%x", md5_gen.Sum(nil))
	result := p.db.Where(repository.ProxyProvider{ProviderId: id}).FirstOrCreate(&repository.ProxyProvider{ProviderId: id, Name: name})
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("provider %s already exists", name))
	}
	return &id, nil
}
func (p ProxyProviderService) ListProvider(ctx context.Context, limit, offset int64) ([]model.ProxyProvider, error) {
	return nil, status.Error(codes.Unimplemented, "ListProvider not implemented")
}
func (p ProxyProviderService) UpdateProvider(ctx context.Context, proxy_provider model.ProxyProvider) error {
	return status.Error(codes.Unimplemented, "UpdateProvider not implemented")
}
func (p ProxyProviderService) DeleteProvider(ctx context.Context, id string) error {
	return status.Error(codes.Unimplemented, "DeleteProvider not implemented")
}
