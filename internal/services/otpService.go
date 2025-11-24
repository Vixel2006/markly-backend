package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"markly/internal/models"
	"markly/internal/repositories"
	"markly/internal/utils"
)

const (
	OTPExpirationMinutes    = 10
	OTPPurposeResetPassword = "reset_password"
)

type OTPService interface {
	GenerateOTPForgotPassword(ctx context.Context, email string) (string, error)
	VerifyOTP(ctx context.Context, email, otpCode string) error
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

	otpCode, err := utils.GenerateSecureOTP(6)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(OTPExpirationMinutes * time.Minute)

	otp := &models.OTP{
		UserID:    user.ID,
		OTPCode:   otpCode,
		Purpose:   OTPPurposeResetPassword,
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

func (s *otpService) VerifyOTP(ctx context.Context, email, otpCode string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	otp, err := s.otpRepo.FindByUserIDAndOTPCode(ctx, user.ID, otpCode, OTPPurposeResetPassword)
	if err != nil {
		return err
	}
	if otp == nil {
		return errors.New("invalid or expired OTP")
	}

	if otp.IsUsed {
		return errors.New("OTP already used")
	}

	if time.Now().After(otp.ExpiresAt) {
		return errors.New("OTP expired")
	}

	err = s.otpRepo.MarkAsUsed(ctx, otp.ID)
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

	otpCode, err := utils.GenerateSecureOTP(6)
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

	subject := "Your One-Time Password"
	body := fmt.Sprintf("Your One-Time Password is: %s", otpCode)
	err = s.emailService.SendEmail(email, subject, body)
	if err != nil {
		return err
	}

	return nil
}
