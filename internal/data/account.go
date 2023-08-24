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
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	tokenString, err := ar.genToken(uint64(u.ID), now)
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
	tokenString, err := ar.genToken(user.Id, now)
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
func (ar *AccountRepo) RefreshToken(ctx context.Context, id uint64) (*v1.LoginResponse, error) {
	user, err := ar.getUserById(ctx, id)
	if err != nil {
		ar.log.Errorf("get user from data error: %s", err.Error())
		return nil, errors.Wrap(err, "GetUserByName error")
	}

	now := time.Now()
	tokenString, err := ar.genToken(user.Id, now)
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
	user := &User{}
	ar.data.db.First(user, "name = ?", name)
	if user.ID == 0 {
		return nil, fmt.Errorf("user with name: %s is not found", name)
	}
	return user.db2pb(), nil
}

func (ar *AccountRepo) getUserById(ctx context.Context, id uint64) (*v1.User, error) {
	user := &User{}
	ar.data.db.First(user, "id = ?", id)
	if user.ID == 0 {
		return nil, fmt.Errorf("user with id: %d is not found", id)
	}
	return user.db2pb(), nil
}

func (ar *AccountRepo) genToken(id uint64, now time.Time) (string, error) {
	claims := &jwtv4.RegisteredClaims{
		ExpiresAt: jwtv4.NewNumericDate(now.Add(ar.expire)),
		Issuer:    "browser",
		Subject:   fmt.Sprintf("%d", id),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	return token.SignedString([]byte(ar.jwtSecret))
}

func (u *User) db2pb() *v1.User {
	createdAt := timestamppb.New(u.CreatedAt)
	updatedAt := timestamppb.New(u.UpdatedAt)
	var deletedAt *timestamppb.Timestamp
	if u.DeletedAt != nil {
		deletedAt = timestamppb.New(*u.DeletedAt)
	}

	return &v1.User{
		Id:       uint64(u.ID),
		Name:     u.Name,
		Password: u.Password,
		Salt:     u.Salt,
		CreateAt: createdAt,
		DeleteAt: deletedAt,
		UpdateAt: updatedAt,
	}
}
