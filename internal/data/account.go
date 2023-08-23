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

type User struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Name      string     `json:"name" gorm:"type:varchar(100);not null"`
	Password  string     `json:"-" gorm:"type:varchar(100)"`
	Salt      string     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

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
	u := &User{
		Name:     username,
		Password: utils.CalcPassword(password, salt),
		Salt:     salt,
	}
	if err := ar.data.db.Save(u).Error; err != nil {
		ar.log.Errorf("save user data error: %s", err.Error())
		return nil, errors.Wrap(err, "save user error")
	}

	now := time.Now()
	tokenString, err := ar.genToken(uint32(u.ID), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       uint64(u.ID),
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
	tokenString, err := ar.genToken(uint32(user.ID), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       uint64(user.ID),
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
	tokenString, err := ar.genToken(uint32(user.ID), now)
	if err != nil {
		ar.log.Errorf("create token error: %s", err.Error())
		return nil, errors.Wrap(err, "create token error")
	}

	return &v1.LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(ar.expire).Unix(),
		Id:       uint64(user.ID),
		Username: user.Name,
	}, nil
}

func (ar *AccountRepo) getUserByName(ctx context.Context, name string) (*User, error) {
	user := &User{}
	ar.data.db.First(user, "name = ?", name)
	if user.ID == 0 {
		return nil, fmt.Errorf("user with name: %s is not found", name)
	}
	return user, nil
}

func (ar *AccountRepo) getUserById(ctx context.Context, id uint32) (*User, error) {
	user := &User{}
	ar.data.db.First(user, "id = ?", id)
	if user.ID == 0 {
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
	return token.SignedString([]byte(ar.jwtSecret))
}
