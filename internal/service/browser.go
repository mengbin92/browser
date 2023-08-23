package service

import (
	"context"

	pb "mengbin92/browser/api/browser/v1"
	"mengbin92/browser/internal/biz"

	klog "github.com/go-kratos/kratos/v2/log"
)

type BrowserService struct {
	account *biz.AccountUsecase
	pb.UnimplementedBrowserServer
}

func NewBrowserService(repo *biz.AccountUsecase, logger klog.Logger) *BrowserService {
	log = klog.NewHelper(logger)
	return &BrowserService{
		account: repo,
	}
}

func (s *BrowserService) GetToken(ctx context.Context, req *pb.Login) (*pb.LoginResponse, error) {
	return s.account.Login(ctx, req.Username, req.Password)
}
func (s *BrowserService) Regisger(ctx context.Context, req *pb.Login) (*pb.LoginResponse, error) {
	return s.account.Register(ctx, req.Username, req.Password)
}
func (s *BrowserService) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.LoginResponse, error) {
	return s.account.RefreshToken(ctx, req.Id)
}
