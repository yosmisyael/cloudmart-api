package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *entity.User) error
	Login(email, password string) (string, string, error)
	RefreshToken(refreshToken string) (string, error)
}

type authService struct {
	repo   repository.AuthRepository
	config *config.Config
}

func NewAuthService(repo repository.AuthRepository, cfg *config.Config) AuthService {
	return &authService{repo: repo, config: cfg}
}

func (s *authService) Register(user *entity.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return errors.New("gagal memproses password")
	}

	user.Password = string(hashedPassword)

	if err := s.repo.Create(user); err != nil {
		return errors.New("email sudah terdaftar")
	}
	return nil
}

func (s *authService) Login(email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", "", errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("email atau password salah")
	}

	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte(s.config.JWTSecret))

	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString([]byte(s.config.JWTSecret))

	_ = s.repo.UpdateRefreshToken(user.ID, refreshToken)

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("refresh token invalid atau expired")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	if user.RefreshToken != tokenString {
		return "", errors.New("refresh token tidak cocok atau sudah di-revoke")
	}

	newAccessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte(s.config.JWTSecret))

	return newAccessToken, nil
}
