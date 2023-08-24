package biz

import (
	"context"
	v1 "mengbin92/browser/api/browser/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AccountUsecase struct {
	repo AccountRepo
}

func NewAccountUsecase(repo AccountRepo, logger log.Logger) *AccountUsecase {
	return &AccountUsecase{repo: repo}
}

func (au *AccountUsecase) Register(ctx context.Context, username, password string) (*v1.LoginResponse, error) {
	return au.repo.Register(ctx, username, password)
}
func (au *AccountUsecase) Login(ctx context.Context, username, password string) (*v1.LoginResponse, error) {
	return au.repo.Login(ctx, username, password)
}
func (au *AccountUsecase) RefreshToken(ctx context.Context, id uint64) (*v1.LoginResponse, error) {
	return au.repo.RefreshToken(ctx, id)
}
