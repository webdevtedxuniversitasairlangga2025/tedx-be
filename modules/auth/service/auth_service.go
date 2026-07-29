package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/auth/dto"
	authRepo "github.com/webdevtedxuniversitasairlangga/modules/auth/repository"
	userDto "github.com/webdevtedxuniversitasairlangga/modules/user/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/user/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/helpers"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error)
	Logout(ctx context.Context, userId string) error
	SendVerificationEmail(ctx context.Context, req userDto.SendVerificationEmailRequest) error
	VerifyEmail(ctx context.Context, req userDto.VerifyEmailRequest) (userDto.VerifyEmailResponse, error)
	SendPasswordReset(ctx context.Context, req dto.SendPasswordResetRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
}

type authService struct {
	userRepository         repository.UserRepository
	refreshTokenRepository authRepo.RefreshTokenRepository
	jwtService             JWTService
	db                     *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo authRepo.RefreshTokenRepository,
	jwtService JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:         userRepo,
		refreshTokenRepository: refreshTokenRepo,
		jwtService:             jwtService,
		db:                     db,
	}
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	user := entities.User{
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		TelpNumber: req.TelpNumber,
		Password:   hashedPassword,
		Role:       "user",
		IsVerified: false,
	}

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	go func() {
		if err := s.SendVerificationEmail(context.Background(), userDto.SendVerificationEmailRequest{
			Email: createdUser.Email,
		}); err != nil {
			log.Printf("failed to send verification email to %s: %v", createdUser.Email, err)
		}
	}()

	return userDto.UserResponse{
		ID:         createdUser.ID.String(),
		Name:       createdUser.Name,
		Email:      createdUser.Email,
		TelpNumber: createdUser.TelpNumber,
		Role:       createdUser.Role,
		IsVerified: createdUser.IsVerified,
	}, nil
}

func (s *authService) Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return dto.TokenResponse{}, userDto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return dto.TokenResponse{}, dto.ErrInvalidCredentials
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	refreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, refreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	refreshToken, err := s.refreshTokenRepository.FindByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return dto.TokenResponse{}, dto.ErrRefreshTokenNotFound
	}

	accessToken := s.jwtService.GenerateAccessToken(refreshToken.UserID.String(), refreshToken.User.Role)
	newRefreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	err = s.refreshTokenRepository.DeleteByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	newRefreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, newRefreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		Role:         refreshToken.User.Role,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userId string) error {
	return s.refreshTokenRepository.DeleteByUserID(ctx, s.db, userId)
}

func (s *authService) SendVerificationEmail(ctx context.Context, req userDto.SendVerificationEmailRequest) error {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return userDto.ErrEmailNotFound
	}

	if user.IsVerified {
		return userDto.ErrAccountAlreadyVerified
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	otp := fmt.Sprintf("%06d", n.Int64())
	expiry := time.Now().Add(15 * time.Minute)

	if err := s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
		"verification_code":   otp,
		"verification_expiry": expiry,
	}).Error; err != nil {
		return err
	}

	subject := "Email Verification"
	renderedBody, err := utils.RenderEmailTemplate("base_email.html", map[string]string{
		"Email": user.Email,
		"Code":  otp,
	})
	if err != nil {
		return err
	}
	return utils.SendMail(user.Email, subject, renderedBody)
}

func (s *authService) VerifyEmail(ctx context.Context, req userDto.VerifyEmailRequest) (userDto.VerifyEmailResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return userDto.VerifyEmailResponse{}, userDto.ErrEmailNotFound
	}

	if user.IsVerified {
		return userDto.VerifyEmailResponse{}, userDto.ErrAccountAlreadyVerified
	}

	if user.VerificationCode != req.Code {
		return userDto.VerifyEmailResponse{}, userDto.ErrTokenInvalid
	}

	if user.VerificationExpiry == nil || time.Now().After(*user.VerificationExpiry) {
		return userDto.VerifyEmailResponse{}, userDto.ErrTokenExpired
	}

	if err := s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
		"is_verified":         true,
		"verification_code":   "",
		"verification_expiry": nil,
	}).Error; err != nil {
		return userDto.VerifyEmailResponse{}, err
	}

	return userDto.VerifyEmailResponse{Email: user.Email, IsVerified: true}, nil
}

func (s *authService) SendPasswordReset(ctx context.Context, req dto.SendPasswordResetRequest) error {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return userDto.ErrEmailNotFound
	}

	resetToken := s.jwtService.GenerateAccessToken(user.ID.String(), "password_reset")

	subject := "Password Reset"
	body := "Please reset your password using this token: " + resetToken

	return utils.SendMail(user.Email, subject, body)
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	token, err := s.jwtService.ValidateToken(req.Token)
	if err != nil || !token.Valid {
		return dto.ErrPasswordResetToken
	}

	userId, err := s.jwtService.GetUserIDByToken(req.Token)
	if err != nil {
		return dto.ErrPasswordResetToken
	}

	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return userDto.ErrUserNotFound
	}

	hashedPassword, err := helpers.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	_, err = s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return err
	}

	return nil
}
