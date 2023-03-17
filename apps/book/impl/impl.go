package impl

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opengoats/goat/app"
	"github.com/opengoats/goat/logger"
	"github.com/opengoats/goat/logger/zap"
	"google.golang.org/grpc"

	"github.com/opengoats/keyauth/apps/book"
	"github.com/opengoats/keyauth/conf"
)

var (
	// Service 服务实例
	svr = &service{}
)

type service struct {
	col *mongo.Collection
	log logger.Logger
	book.UnimplementedServiceServer
}

func (s *service) Config() error {

	db, err := conf.C().Mongo.GetDB()
	if err != nil {
		return err
	}
	s.col = db.Collection(s.Name())

	s.log = zap.L().Named(s.Name())
	return nil
}

func (s *service) Name() string {
	return book.AppName
}

func (s *service) Registry(server *grpc.Server) {
	book.RegisterServiceServer(server, svr)
}

func init() {
	app.RegistryGrpcApp(svr)
}
