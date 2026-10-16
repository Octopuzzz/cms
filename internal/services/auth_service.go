// Package services provides JWT authentication business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication operations
type AuthService struct {
	db  *gorm.DB
	cfg *config.JWTConfig
}

// Claims represents JWT claims
type Claims struct {
	UserID       uint     `json:"user_id"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	Roles        []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenPair holds both tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LoginRequest for login endpoint
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest for registration
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// NewAuthService creates a new AuthService
func NewAuthService(db *gorm.DB, cfg *config.JWTConfig) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	var existing models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hash),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	// Assign default viewer role
	var viewerRole models.Role
	if err := s.db.Where("name = ?", "viewer").First(&viewerRole).Error; err == nil {
		s.db.Create(&models.UserRole{UserID: user.ID, RoleID: viewerRole.ID})
	}

	user.Password = ""
	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*TokenPair, *models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, nil, errors.New("account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	tokens, err := s.generateTokenPair(ctx, &user)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)
	user.Password = ""
	return tokens, &user, nil
}

// RefreshTokens issues a new token pair from a valid refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	var stored models.RefreshToken
	if err := s.db.Preload("User.Roles").Where("token = ? AND is_revoked = false AND expires_at > ?", refreshToken, time.Now()).First(&stored).Error; err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Revoke old token
	s.db.Model(&stored).Update("is_revoked", true)

	return s.generateTokenPair(ctx, stored.User)
}

// ValidateToken parses and validates a JWT token
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User) (*TokenPair, error) {
	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	expiresAt := time.Now().Add(s.cfg.AccessTokenExpire)
	claims := &Claims{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		IsSuperAdmin: user.IsSuperAdmin,
		Roles:        roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshTokenStr := uuid.New().String() + uuid.New().String()
	refreshExpiry := time.Now().Add(s.cfg.RefreshTokenExpire)
	s.db.Create(&models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: refreshExpiry,
	})

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    expiresAt,
	}, nil
}

// GetUserByID retrieves full user details with roles
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}
