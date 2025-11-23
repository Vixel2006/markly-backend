package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"markly/internal/models"
	"markly/internal/repositories"
	"markly/internal/utils"
)

const (
	OTPExpirationMinutes     = 10
	OTPPurposeForgotPassword = "forgot_password"
)

type OTPService interface {
	GenerateOTPForgotPassword(ctx context.Context, email string) (string, error)
	VerifyOTPForgotPassword(ctx context.Context, email, otpCode string) (*models.User, error)
	ResetPassword(ctx context.Context, userID, newPassword string) error
	SendOTP(ctx context.Context, email string) error
}

type otpService struct {
	userRepo     repositories.UserRepository
	otpRepo      repositories.OTPRepository
	emailService EmailService
}

func NewOTPService(userRepo repositories.UserRepository, otpRepo repositories.OTPRepository, emailService EmailService) OTPService {
	return &otpService{userRepo: userRepo, otpRepo: otpRepo, emailService: emailService}
}

func (s *otpService) GenerateOTPForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	otpCode, err := generateSecureOTP(6)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(OTPExpirationMinutes * time.Minute)

	otp := &models.OTP{
		UserID:    user.ID,
		OTPCode:   otpCode,
		Purpose:   OTPPurposeForgotPassword,
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}

	_, err = s.otpRepo.Create(ctx, otp)
	if err != nil {
		return "", err
	}

	subject := "Your Password Reset OTP"
	body := fmt.Sprintf("Your OTP for password reset is: %s", otpCode)
	err = s.emailService.SendEmail(email, subject, body)
	if err != nil {
		return "", err
	}

	return otpCode, nil
}

func generateSecureOTP(length int) (string, error) {
	const otpChars = "0123456789"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

func (s *otpService) VerifyOTPForgotPassword(ctx context.Context, email, otpCode string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	otp, err := s.otpRepo.FindByUserIDAndOTPCode(ctx, user.ID, otpCode, OTPPurposeForgotPassword)
	if err != nil {
		return nil, err
	}
	if otp == nil {
		return nil, errors.New("invalid or expired OTP")
	}

	if otp.IsUsed {
		return nil, errors.New("OTP already used")
	}

	if time.Now().After(otp.ExpiresAt) {
		return nil, errors.New("OTP expired")
	}

	err = s.otpRepo.MarkAsUsed(ctx, otp.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *otpService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	updateFields := bson.M{
		"password":   user.Password,
		"updated_at": user.UpdatedAt,
	}

	_, err = s.userRepo.Update(ctx, user.ID, updateFields)
	if err != nil {
		return err
	}

	return nil
}

func (s *otpService) SendOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	otpCode, err := generateSecureOTP(6)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(OTPExpirationMinutes * time.Minute)

	otp := &models.OTP{
		UserID:    user.ID,
		OTPCode:   otpCode,
		Purpose:   "generic_otp",
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}

	_, err = s.otpRepo.Create(ctx, otp)
	if err != nil {
		return err
	}

	// Send OTP via email
	subject := "Your One-Time Password"
	body := fmt.Sprintf("Your One-Time Password is: %s", otpCode)
	err = s.emailService.SendEmail(email, subject, body)
	if err != nil {
		return err
	}

	return nil
}
