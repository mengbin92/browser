package data

import (
	"fmt"
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mengbin92/browser/db"
	"github.com/mengbin92/browser/utils"
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

var (
	securityKey = []byte{
		0x46, 0x68, 0x43, 0x6e, 0x77, 0x4f, 0x44, 0x74,
		0x62, 0x51, 0x52, 0x6a, 0x37, 0x73, 0x84, 0x79,
		0x75, 0x4f, 0x65, 0x6d, 0x57, 0x62, 0x74, 0x6a,
		0x69, 0x75, 0x63, 0x6d, 0x43, 0x43, 0x54, 0x42,
		0x6e, 0x62, 0x47, 0x6a, 0x70, 0x34,
	}
)

type LoginResponse struct {
	Token    string `json:"token,omitempty"`
	Expire   int64  `json:"expire,omitempty"`
	ID       uint   `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

func GetUserByName(name string) (*User, error) {
	var user User
	db.Get().First(&user, "name = ?", name)
	if user.ID == 0 {
		return nil, fmt.Errorf("user with name: %s is not found", name)
	}
	return &user, nil
}

func GetUserByID(id uint) (*User, error) {
	var user User
	db.Get().First(&user, "id = ?", id)
	if user.ID == 0 {
		return nil, fmt.Errorf("user with id: %d is not found", id)
	}
	return &user, nil
}

func genSalt() string {
	uid, _ := uuid.NewRandom()
	return uid.String()
}

func RegisterUser(name, password string) (*LoginResponse, error) {
	salt := genSalt()
	u := &User{
		Name:     name,
		Password: utils.CalcPassword(password, salt),
		Salt:     salt,
	}
	if err := db.Get().Save(u).Error; err != nil {
		return nil, errors.Wrap(err, "RegisterUser error")
	}

	now := time.Now()
	claims := &jwtv5.RegisteredClaims{
		ExpiresAt: jwtv5.NewNumericDate(now.Add(30 * time.Minute)),
		Issuer:    "browser",
		Subject:   fmt.Sprintf("%d", u.ID),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(securityKey)
	if err != nil {
		return nil, errors.Wrap(err, "create token error")
	}

	return &LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(30 * time.Minute).Unix(),
		ID:       u.ID,
		Username: u.Name,
	}, nil
}

func Login(name, password string) (*LoginResponse, error) {
	user, err := GetUserByName(name)
	if err != nil {
		return nil, errors.Wrap(err, "GetUserByName error")
	}

	if utils.CalcPassword(password, user.Salt) != user.Password {
		return nil, errors.New("user name or password is incorrect")
	}

	now := time.Now()
	claims := &jwtv5.RegisteredClaims{
		ExpiresAt: jwtv5.NewNumericDate(now.Add(30 * time.Minute)),
		Issuer:    "browser",
		Subject:   fmt.Sprintf("%d", user.ID),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(securityKey)
	if err != nil {
		return nil, errors.Wrap(err, "create token error")
	}

	return &LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(30 * time.Minute).Unix(),
		ID:       user.ID,
		Username: user.Name,
	}, nil
}

func RefreshToken(id uint) (*LoginResponse, error) {
	user, err := GetUserByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "GetUserByName error")
	}

	now := time.Now()
	claims := &jwtv5.RegisteredClaims{
		ExpiresAt: jwtv5.NewNumericDate(now.Add(30 * time.Minute)),
		Issuer:    "browser",
		Subject:   fmt.Sprintf("%d", user.ID),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(securityKey)
	if err != nil {
		return nil, errors.Wrap(err, "create token error")
	}
	return &LoginResponse{
		Token:    tokenString,
		Expire:   now.Add(30 * time.Minute).Unix(),
		ID:       user.ID,
		Username: user.Name,
	}, nil
}

func ParseJWT(tokenString string) error {
	token, err := jwtv5.Parse(tokenString, func(token *jwtv5.Token) (interface{}, error) {
		return securityKey, nil
	})
	if err != nil {
		return errors.Wrap(err, "new token parse function error")
	}
	// 验证 JWT 的有效性
	if !token.Valid {
		return errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwtv5.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return errors.Wrap(err, "GetExpirationTime from token error")
	}

	if time.Until(exp.Time) > 0 {
		return errors.New("the token expires")
	}

	id, err := strconv.Atoi(claims["sub"].(string))
	if err != nil {
		return errors.Wrapf(err, `strconv.Atoi with claims["sub"]: %v`, claims["sub"])
	}

	var user User
	if err := db.Get().First(&user, "id = ?", id).Error; err != nil {
		return errors.Wrap(err, "get user by id error")
	}
	return nil
}
