package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/infrastructure/pkg/mailer"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/jwt"
	"github.com/mrbayss/Go-Commerce/internal/pkg/otp"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var dummyBcryptHash string

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-login-salt-for-constant-time"), bcrypt.DefaultCost)
	if err == nil {
		dummyBcryptHash = string(hash)
	}
}

type AuthService interface {
	SignUp(ctx context.Context, request model.SignUpReq) (*model.SignUpRes, error)
	VerifyOTP(ctx context.Context, request model.VerifyOTP) (*model.TokenPairRes, error)
	ResendOTP(ctx context.Context, request model.ResendOTPReq) (*model.ResendOTPRes, error)
	Login(ctx context.Context, request model.LoginReq) (*model.TokenPairRes, error)
}

type authService struct {
	AuthRepo         repository.AuthRepository
	Log              *logrus.Logger
	Redis            *redis.Client
	Validation       *validator.Validate
	Mail             *mailer.Client
	OTPTtl           time.Duration
	SignupRateWindow time.Duration
	ResendCooldown      time.Duration
	MaxSignupPerHour    int
	MaxSignupPerDay     int
	MaxResendPerSession int
	LoginMaxAttempts    int
	LoginLockoutTime time.Duration
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string
}

type signupPending struct {
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Phone        *string `json:"phone,omitempty"`
}

func NewAuthService(
	authRepo repository.AuthRepository,
	log *logrus.Logger,
	redis *redis.Client,
	validation *validator.Validate,
	mail *mailer.Client,
	otpttl time.Duration,
	signupratewindow time.Duration,
	resendcooldown time.Duration,
	maxsignupperhour int,
	maxsignupperday int,
	maxresendpersession int,
	loginMaxAttempts int,
	loginLockoutTime time.Duration,
	jwtSecret string,
	jwtAccessExpiry time.Duration,
	jwtRefreshExpiry time.Duration,
	jwtIssuer string,
) AuthService {
	if maxresendpersession <= 0 {
		maxresendpersession = 3
	}
	return &authService{
		AuthRepo:            authRepo,
		Log:                 log,
		Redis:               redis,
		Validation:          validation,
		Mail:                mail,
		OTPTtl:              otpttl,
		SignupRateWindow:    signupratewindow,
		ResendCooldown:      resendcooldown,
		MaxSignupPerHour:    maxsignupperhour,
		MaxSignupPerDay:     maxsignupperday,
		MaxResendPerSession: maxresendpersession,
		LoginMaxAttempts:    loginMaxAttempts,
		LoginLockoutTime:    loginLockoutTime,
		JWTSecret:           jwtSecret,
		JWTAccessExpiry:     jwtAccessExpiry,
		JWTRefreshExpiry:    jwtRefreshExpiry,
		JWTIssuer:           jwtIssuer,
	}
}

func (as *authService) SignUp(ctx context.Context, request model.SignUpReq) (*model.SignUpRes, error) {
	log, email := as.getLoggerAndEmail(ctx, "SignUp", request.Email)

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	exists, err := as.AuthRepo.Exists(ctx, "email", email)
	if err != nil {
		log.WithError(err).Error("failed to lookup user")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to lookup user")
	}
	if exists {
		log.Error("email already registered")
		return nil, apperrors.Wrap(nil, http.StatusConflict, "email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error("failed to hash password")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to hashing password")
	}
	hashedPassword := string(hash)

	var phone *string
	if request.Phone != "" {
		phone = &request.Phone
	}

	pending := signupPending{
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Phone:        phone,
	}

	pendingBytes, err := json.Marshal(pending)
	if err != nil {
		log.WithError(err).Error("failed to marshal pending signup")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to store signup session")
	}

	codeOTP, err := otp.Generate()
	if err != nil {
		log.WithError(err).Error("failed to generate otp")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed generated otp")
	}

	hashOTP := otp.Hash(codeOTP, as.JWTSecret, email)

	keys := signupKeysFor(email)
	if err := as.Redis.Set(ctx, keys.Pending, string(pendingBytes), as.OTPTtl).Err(); err != nil {
		log.WithError(err).Error("failed to store pending signup in redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store signup session")
	}
	if err := as.Redis.Set(ctx, keys.OTP, hashOTP, as.OTPTtl).Err(); err != nil {
		log.WithError(err).Error("failed to store otp in redis")
		_ = as.Redis.Del(ctx, keys.Pending).Err()
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store otp")
	}
	_ = as.Redis.Del(ctx, keys.Attempts, keys.ResendCount).Err()

	if err := as.Mail.SendTemplate(ctx, []string{email}, "GoCommerce OTP Verify", "otp_signup.html", mailer.OtpData{
		Code:    codeOTP,
		Expired: int(as.OTPTtl.Minutes()),
	}); err != nil {
		log.WithError(err).Error("failed to send otp email")
		_ = as.Redis.Del(ctx, keys.Pending, keys.OTP, keys.Attempts, keys.ResendCount).Err()
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to send otp to email")
	}

	res := model.SignUpRes{
		Email:     email,
		ExpiredAt: time.Now().Add(as.OTPTtl),
	}

	return &res, nil
}

func (as *authService) VerifyOTP(ctx context.Context, request model.VerifyOTP) (*model.TokenPairRes, error) {
	log, email := as.getLoggerAndEmail(ctx, "VerifyOTP", request.Email)

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	keys := signupKeysFor(email)

	storedHash, err := as.Redis.Get(ctx, keys.OTP).Result()
	if err == redis.Nil {
		log.Error("otp expired or not found")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "OTP expired or invalid")
	}
	if err != nil {
		log.WithError(err).Error("failed to get otp from redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to verify otp")
	}

	if !otp.Verify(request.CodeOTP, storedHash, as.JWTSecret, email) {
		attempts, incrErr := as.Redis.Incr(ctx, keys.Attempts).Result()
		if incrErr != nil {
			log.WithError(incrErr).Error("failed to increment otp attempt counter")
			return nil, apperrors.Wrap(incrErr, http.StatusBadGateway, "failed to verify otp")
		}
		if attempts == 1 {
			_ = as.Redis.Expire(ctx, keys.Attempts, as.OTPTtl).Err()
		}
		if attempts >= 5 {
			log.Error("otp attempt limit exceeded")
			_ = as.Redis.Del(ctx, keys.Pending, keys.OTP, keys.Attempts, keys.ResendCount).Err()
			return nil, apperrors.Wrap(nil, http.StatusTooManyRequests, "too many invalid OTP attempts")
		}

		log.Error("invalid otp code")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid OTP code")
	}

	consumedHash, err := as.Redis.GetDel(ctx, keys.OTP).Result()
	if err == redis.Nil || consumedHash != storedHash {
		log.Error("otp already used or expired")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "OTP expired or invalid")
	}
	if err != nil {
		log.WithError(err).Error("failed to consume otp")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to verify otp")
	}

	pendingData, err := as.Redis.Get(ctx, keys.Pending).Result()
	if err == redis.Nil {
		log.Error("pending signup expired or not found")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "signup session expired")
	}
	if err != nil {
		log.WithError(err).Error("failed to get pending signup from redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to verify otp")
	}

	var pending signupPending
	if err := json.Unmarshal([]byte(pendingData), &pending); err != nil {
		log.WithError(err).Error("failed to unmarshal pending signup")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to process signup")
	}

	user := &entity.User{
		Email:        pending.Email,
		PasswordHash: &pending.PasswordHash,
		FirstName:    pending.FirstName,
		LastName:     pending.LastName,
		Phone:        pending.Phone,
		Role:         entity.UserRoleBuyer,
		IsActive:     true,
		IsVerified:   true,
	}

	User, err := as.AuthRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Error("email already registered")
			_ = as.Redis.Del(ctx, keys.Pending, keys.OTP, keys.Attempts, keys.ResendCount).Err()
			return nil, apperrors.Wrap(err, http.StatusConflict, "email already registered")
		}
		log.WithError(err).Error("failed to create user")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create user")
	}

	sessionID := uuid.New().String()
	tokenPair, err := jwt.GenerateTokenPair(
		sessionID,
		User.ID.String(),
		User.Email,
		string(User.Role),
		User.IsVerified,
		as.JWTSecret,
		as.JWTAccessExpiry,
		as.JWTRefreshExpiry,
		as.JWTIssuer,
	)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to generate tokens")
	}

	_ = as.Redis.Del(ctx, keys.Pending, keys.OTP, keys.Attempts, keys.ResendCount).Err()

	refreshTokenHash := jwt.HashToken(tokenPair.RefreshToken)
	sessionKey := fmt.Sprintf("session:%s:%s", User.ID.String(), sessionID)
	if err := as.Redis.Set(ctx, sessionKey, refreshTokenHash, as.JWTRefreshExpiry).Err(); err != nil {
		log.WithError(err).Error("failed to store session in redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store session")
	}

	res := &model.TokenPairRes{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return res, nil
}

func (as *authService) ResendOTP(ctx context.Context, request model.ResendOTPReq) (*model.ResendOTPRes, error) {
	log, email := as.getLoggerAndEmail(ctx, "ResendOTP", request.Email)

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	keys := signupKeysFor(email)

	exists, err := as.Redis.Exists(ctx, keys.Pending).Result()
	if err != nil {
		log.WithError(err).Error("failed to check pending signup")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to check signup session")
	}
	if exists == 0 {
		log.Error("no pending signup found")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "no pending signup found, please signup first")
	}

	resendCount, err := as.Redis.Get(ctx, keys.ResendCount).Int()
	if err != nil && err != redis.Nil {
		log.WithError(err).Error("failed to check resend count from redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to check resend limit")
	}
	if resendCount >= as.MaxResendPerSession {
		log.Warn("max resend attempts reached for this session")
		_ = as.Redis.Del(ctx, keys.Pending, keys.OTP, keys.Attempts, keys.ResendCount).Err()
		return nil, apperrors.Wrap(nil, http.StatusTooManyRequests, "maximum resend attempts reached for this session, please signup again")
	}

	codeOTP, err := otp.Generate()
	if err != nil {
		log.WithError(err).Error("failed to generate otp")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed generated otp")
	}

	hashOTP := otp.Hash(codeOTP, as.JWTSecret, email)

	pipe := as.Redis.Pipeline()
	pipe.Set(ctx, keys.OTP, hashOTP, as.OTPTtl)
	pipe.Expire(ctx, keys.Pending, as.OTPTtl)
	pipe.Incr(ctx, keys.ResendCount)
	pipe.Expire(ctx, keys.ResendCount, as.OTPTtl)
	pipe.Del(ctx, keys.Attempts)
	if _, err := pipe.Exec(ctx); err != nil {
		log.WithError(err).Error("failed to store otp and update session in redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store otp")
	}

	if err := as.Mail.SendTemplate(ctx, []string{email}, "GoCommerce OTP Verify", "otp_signup.html", mailer.OtpData{
		Code:    codeOTP,
		Expired: int(as.OTPTtl.Minutes()),
	}); err != nil {
		log.WithError(err).Error("failed to send otp email")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to send otp to email")
	}

	res := model.ResendOTPRes{
		Email:     email,
		ExpiredAt: time.Now().Add(as.OTPTtl),
	}

	return &res, nil
}

func (as *authService) Login(ctx context.Context, request model.LoginReq) (*model.TokenPairRes, error) {
	log, email := as.getLoggerAndEmail(ctx, "Login", request.Email)

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	user, err := as.AuthRepo.GetByColumn(ctx, "email", email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(request.Password))
			log.Warn("login failed: user not found")
			return nil, apperrors.Wrap(nil, http.StatusBadRequest, "email or password is wrong")
		}
		log.WithError(err).Error("failed to query user from database")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "internal server error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(request.Password)); err != nil {
		log.WithError(err).Error("email or password is wrong")
		return nil, apperrors.Wrap(err, http.StatusBadRequest, "email or password is wrong")
	}

	if !user.IsActive {
		return nil, apperrors.Wrap(nil, http.StatusForbidden, "account is inactive")
	}
	if !user.IsVerified {
		return nil, apperrors.Wrap(nil, http.StatusForbidden, "account is not verified")
	}

	sessionID := uuid.New().String()
	tokenPair, err := jwt.GenerateTokenPair(
		sessionID,
		user.ID.String(),
		user.Email,
		string(user.Role),
		user.IsVerified,
		as.JWTSecret,
		as.JWTAccessExpiry,
		as.JWTRefreshExpiry,
		as.JWTIssuer,
	)
	if err != nil {
		log.WithError(err).Error("failed to generate token")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to generate token")
	}

	refreshTokenHash := jwt.HashToken(tokenPair.RefreshToken)
	sessionKey := fmt.Sprintf("session:%s:%s", user.ID.String(), sessionID)
	if err := as.Redis.Set(ctx, sessionKey, refreshTokenHash, as.JWTRefreshExpiry).Err(); err != nil {
		log.WithError(err).Error("failed to store session in redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store session")
	}

	res := &model.TokenPairRes{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return res, nil
}

type signupKeys struct {
	Pending     string
	OTP         string
	Attempts    string
	ResendCount string
}

func signupKeysFor(email string) signupKeys {
	emailLower := strings.ToLower(strings.TrimSpace(email))
	return signupKeys{
		Pending:     "signup:pending:" + emailLower,
		OTP:         "signup:otp:" + emailLower,
		Attempts:    "signup:attempts:" + emailLower,
		ResendCount: "signup:resend:" + emailLower,
	}
}

func (as *authService) getLoggerAndEmail(ctx context.Context, functionName string, email string) (*logrus.Entry, string) {
	emaill := strings.ToLower(strings.TrimSpace(email))
	reqID := ctx.Value("request_id").(string)
	return as.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
		"email":      emaill,
	}), emaill
}
