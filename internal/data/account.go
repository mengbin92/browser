package data

import (
	"context"
	"fmt"
	v1 "mengbin92/browser/api/browser/v1"
	"mengbin92/browser/internal/biz"
	"mengbin92/browser/internal/conf"
	"mengbin92/browser/internal/utils"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type AccountRepo struct {
	data      *Data
	log       *log.Helper
	jwtSecret string
	expire    time.Duration
}

func NewAccountRepo(data *Data, logger log.Logger, auth *conf.Auth) biz.AccountRepo {
	return &AccountRepo{
		data:      data,
		log:       log.NewHelper(logger),
		jwtSecret: auth.JwtSecret,
		expire:    auth.Expire.AsDuration(),
	}
}

func genSalt() string {
	uid, _ := uuid.NewRandom()
	return uid.String()
}

func (ar *AccountRepo) Register(ctx context.Context, username, password string) (*v1.LoginResponse, error) {
	salt := genSalt()
	u := &v1.User{
		Name:     username,
		Password: utils.CalcPassword(password, salt),
		Salt:     salt,
	}
	if err := ar.data.db.Save(u).Error; err != nil {
		ar.log.Errorf("save user data error: %s", err.Error())
		return nil, errors.Wrap(err, "save user error")
	}

	now := time.Now()
	tokenString, err := ar.genToken(uint32(u.Id), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       u.Id,
		Username: u.Name,
	}, nil
}
func (ar *AccountRepo) Login(ctx context.Context, username, password string) (*v1.LoginResponse, error) {
	user, err := ar.getUserByName(ctx, username)
	if err != nil {
		ar.log.Errorf("get user from data error: %s", err.Error())
		return nil, errors.Wrap(err, "GetUserByName error")
	}

	if utils.CalcPassword(password, user.Salt) != user.Password {
		ar.log.Error("user name or password is incorrect")
		return nil, errors.New("user name or password is incorrect")
	}
	now := time.Now()
	tokenString, err := ar.genToken(uint32(user.Id), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       user.Id,
		Username: user.Name,
	}, nil
}
func (ar *AccountRepo) RefreshToken(ctx context.Context, id uint32) (*v1.LoginResponse, error) {
	user, err := ar.getUserById(ctx, id)
	if err != nil {
		ar.log.Errorf("get user from data error: %s", err.Error())
		return nil, errors.Wrap(err, "GetUserByName error")
	}

	now := time.Now()
	tokenString, err := ar.genToken(uint32(user.Id), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       user.Id,
		Username: user.Name,
	}, nil
}

func (ar *AccountRepo) getUserByName(ctx context.Context, name string) (*v1.User, error) {
	user := &v1.User{}
	ar.data.db.First(user, "name = ?", name)
	if user.Id == 0 {
		return nil, fmt.Errorf("user with name: %s is not found", name)
	}
	return user, nil
}

func (ar *AccountRepo) getUserById(ctx context.Context, id uint32) (*v1.User, error) {
	user := &v1.User{}
	ar.data.db.First(user, "id = ?", id)
	if user.Id == 0 {
		return nil, fmt.Errorf("user with id: %d is not found", id)
	}
	return user, nil
}

func (ar *AccountRepo) genToken(id uint32, now time.Time) (string, error) {
	claims := &jwtv5.RegisteredClaims{
		ExpiresAt: jwtv5.NewNumericDate(now.Add(ar.expire)),
		Issuer:    "browser",
		Subject:   fmt.Sprintf("%d", id),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(ar.jwtSecret)
}
