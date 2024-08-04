package service

// Proxy provides operations on proxy.
import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	event "github.com/WALL-EEEEEEE/proxy-service/manager/event"
	"github.com/WALL-EEEEEEE/proxy-service/manager/model"
	"github.com/WALL-EEEEEEE/proxy-service/manager/repository"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type IProxyApiService interface {
	GetApi(context.Context, string) (*model.ProxyApi, error)
	GetApiByProvider(context.Context, string) ([]model.ProxyApi, error)
	AddApi(context.Context, model.ProxyApi) (*string, error)
	UpdateApi(context.Context, model.ProxyApi) error
	DeleteApi(context.Context, string) error
}

type ProxyApiService struct {
	db        *gorm.DB
	event_bus *redis.Client
}

func exists[T any](db *gorm.DB, condition T) (bool, error) {
	var repo T
	result := db.Where(condition).First(&repo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func NewProxyApiService(db *gorm.DB, event_bus *redis.Client) *ProxyApiService {
	return &ProxyApiService{
		db:        db,
		event_bus: event_bus,
	}
}

func (p ProxyApiService) GetApi(ctx context.Context, id string) (*model.ProxyApi, error) {
	var api repository.ProxyApi
	result := p.db.Where(repository.ProxyApi{ApiId: id}).First(&api)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("proxy api %s not exists", id))
	}
	ret_api := model.ProxyApi{}
	ret_api.Id = api.ApiId
	ret_api.UpdateInterval = api.UpdateInterval
	ret_api.Name = api.Name
	ret_api.ProviderId = api.ProviderId
	ret_api.Service = model.Service{Host: api.Service.Host, Name: api.Service.Name, Params: api.Service.Params}
	return &ret_api, nil
}

func (p ProxyApiService) GetApiByProvider(ctx context.Context, provider_id string) ([]model.ProxyApi, error) {
	is_exists, err := exists[repository.ProxyProvider](p.db, repository.ProxyProvider{ProviderId: provider_id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !is_exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("proxy provider %s not exists", provider_id))
	}
	var apis []repository.ProxyApi
	result := p.db.Where(repository.ProxyApi{ProviderId: provider_id}).Find(&apis)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if result.RowsAffected == 0 {
		return []model.ProxyApi{}, nil
	}
	var ret_apis []model.ProxyApi
	for _, api := range apis {
		ret_api := model.ProxyApi{}
		ret_api.Id = api.ApiId
		ret_api.UpdateInterval = api.UpdateInterval
		ret_api.Name = api.Name
		ret_api.ProviderId = api.ProviderId
		ret_api.Service = model.Service{Host: api.Service.Host, Name: api.Service.Name, Params: api.Service.Params}
		ret_apis = append(ret_apis, ret_api)
	}
	return ret_apis, nil
}

func (p ProxyApiService) AddApi(ctx context.Context, proxy_api model.ProxyApi) (*string, error) {
	//Check the existence of Provider
	provider_id := proxy_api.ProviderId
	is_exists, err := exists[repository.ProxyProvider](p.db, repository.ProxyProvider{ProviderId: proxy_api.ProviderId})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !is_exists {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("proxy provider %s not exists", provider_id))
	}
	//Insert the api under provider
	md5_gen := md5.New()
	_, err = md5_gen.Write([]byte(proxy_api.Name + "_" + provider_id))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	id := fmt.Sprintf("%x", md5_gen.Sum(nil))
	if err != nil {
		return nil, err
	}
	var api repository.ProxyApi = repository.ProxyApi{
		ProviderId:     provider_id,
		Name:           proxy_api.Name,
		ApiId:          id,
		Service:        repository.Service{Host: proxy_api.Service.Host, Name: proxy_api.Service.Name, Params: proxy_api.Service.Params},
		UpdateInterval: proxy_api.UpdateInterval,
	}
	result := p.db.Where(repository.ProxyApi{ApiId: id}).FirstOrCreate(&api)
	if result.Error != nil {
		return nil, status.Error(codes.Unknown, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("proxy api %s already exists under proxy provider %s", proxy_api.Name, provider_id))
	}
	var check_api_item = model.ProxyApi{
		Id:             id,
		UpdateInterval: proxy_api.UpdateInterval,
		Name:           proxy_api.Name,
		Service:        model.Service{Host: proxy_api.Service.Host, Name: proxy_api.Service.Name, Params: proxy_api.Service.Params},
		ProviderId:     provider_id,
	}
	check_api_item_data, err := json.Marshal(check_api_item)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	retCmd := p.event_bus.XAdd(ctx, &redis.XAddArgs{
		Stream: event.EVENT_CHECKAPI_CREATED,
		Values: map[string]interface{}{
			"data": check_api_item_data,
		},
	})
	err = retCmd.Err()
	if err != nil {
		log.Warnf("Failed check_api (id: %s) event failed: %s (code: %s)", id, err.Error(), retCmd.Val())
	}
	return &id, nil
}
func (p ProxyApiService) UpdateApi(ctx context.Context, proxy_provider model.ProxyApi) error {
	return nil
}
func (p ProxyApiService) DeleteApi(ctx context.Context, id string) error {
	return nil
}
