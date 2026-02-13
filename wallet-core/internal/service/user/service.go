package user

import (
	"context"
	"errors"
	"time"

	"wallet-core/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("username or email already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Register 创建新用户
func (s *Service) Register(ctx context.Context, username, email, password string) (int64, error) {
	// 1. Hash Password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// 2. Create User Model
	user := model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPwd),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 3. Insert to DB
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, ErrUserAlreadyExists
		}
		return 0, err
	}

	return int64(user.ID), nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, email, password string) (string, int64, string, error) {
	var user model.User
	// 1. Find User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, "", ErrUserNotFound
		}
		return "", 0, "", err
	}

	// 2. Compare Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, "", ErrInvalidPassword
	}

	// 3. Generate Token (Simple UUID for Phase 1)
	token := uuid.New().String()

	return token, int64(user.ID), user.Username, nil
}

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
