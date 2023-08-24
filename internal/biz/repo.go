package biz

import (
	"context"
	v1 "mengbin92/browser/api/browser/v1"
)

type AccountRepo interface {
	Register(context.Context, string, string) (*v1.LoginResponse, error)
	Login(context.Context, string, string) (*v1.LoginResponse, error)
	RefreshToken(context.Context, uint64) (*v1.LoginResponse, error)
}
