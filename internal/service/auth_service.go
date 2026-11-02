package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
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
)

type AuthService interface {
	SignUp(ctx context.Context, request model.SignUpReq) (*model.SignUpRes, error)
	VerifyOTP(ctx context.Context, request model.VerifyOTP) (*model.VerifyOTPRes, error)
	ResendOTP(ctx context.Context, request model.ResendOTPReq) (*model.ResendOTPRes, error)
}

type authService struct {
	AuthRepo         repository.AuthRepository
	Log              *logrus.Logger
	Redis            *redis.Client
	Validation       *validator.Validate
	Mail             *mailer.Client
	OTPTtl           time.Duration
	SignupRateWindow time.Duration
	ResendCooldown   time.Duration
	MaxSignupPerHour int
	MaxSignupPerDay  int
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
	jwtSecret string,
	jwtAccessExpiry time.Duration,
	jwtRefreshExpiry time.Duration,
	jwtIssuer string,
) AuthService {
	return &authService{
		AuthRepo:         authRepo,
		Log:              log,
		Redis:            redis,
		Validation:       validation,
		Mail:             mail,
		OTPTtl:           otpttl,
		SignupRateWindow: signupratewindow,
		ResendCooldown:   resendcooldown,
		MaxSignupPerHour: maxsignupperhour,
		MaxSignupPerDay:  maxsignupperday,
		JWTSecret:        jwtSecret,
		JWTAccessExpiry:  jwtAccessExpiry,
		JWTRefreshExpiry: jwtRefreshExpiry,
		JWTIssuer:        jwtIssuer,
	}
}

const rateLimitScript = `
	local email = KEYS[1]
	local minuteKey = KEYS[2]
	local hourKey = KEYS[3]
	local dayKey = KEYS[4]

	local minuteWindow = tonumber(ARGV[1])
	local hourMax      = tonumber(ARGV[2])
	local dayMax       = tonumber(ARGV[3])
	local hourWindowSec = tonumber(ARGV[4])

	-- Layer 1: Minute cooldown - if key exists, reject
	if redis.call("EXISTS", minuteKey) == 1 then
		return "minute_cooldown_active"
	end

	-- Set minute cooldown with TTL
	redis.call("SET", minuteKey, "1", "EX", minuteWindow)

	-- Layer 2: Hour counter
	local hourCount = redis.call("INCR", hourKey)
	if hourCount == 1 then
		redis.call("EXPIRE", hourKey, hourWindowSec)
	end
	if hourCount > hourMax then
		return "hour_limit_exceeded: " .. hourCount .. "/" .. hourMax
	end

	-- Layer 3: Day counter
	local dayCount = redis.call("INCR", dayKey)
	if dayCount == 1 then
		redis.call("EXPIRE", dayKey, 86400)
	end
	if dayCount > dayMax then
		return "day_limit_exceeded: " .. dayCount .. "/" .. dayMax
	end

	return nil
`

func (as *authService) checkRateLimit(ctx context.Context, email string, minuteWindow time.Duration, hourWindowSec int) error {
	log := as.Log.WithFields(logrus.Fields{
		"request_id": ctx.Value("request_id"),
		"function":   "Auth.checkRateLimit",
		"email":      email,
	})

	emailLower := strings.ToLower(strings.TrimSpace(email))

	minuteKey := "rl:rate:signup:minute:" + emailLower
	hourKey := "rl:rate:signup:hour:" + emailLower
	dayKey := "rl:rate:signup:day:" + emailLower

	result, err := as.Redis.Eval(ctx, rateLimitScript, []string{
		emailLower, minuteKey, hourKey, dayKey,
	}, int(minuteWindow.Seconds()), hourWindowSec, as.MaxSignupPerDay, int(as.SignupRateWindow.Seconds())).Result()
	if err != nil && err != redis.Nil {
		log.WithError(err).Error("failed to execute rate limit lua script")
		return apperrors.Wrap(err, http.StatusBadGateway, "failed to check rate limit")
	}

	if err == redis.Nil || result == nil {
		return nil
	}

	errorMsg, ok := result.(string)
	if !ok {
		log.Error("unexpected rate limit script result type")
		return apperrors.Wrap(nil, http.StatusInternalServerError, "failed to check rate limit")
	}

	log.WithField("error", errorMsg).Error("rate limit exceeded")
	return apperrors.Wrap(nil, http.StatusTooManyRequests, "rate limit exceeded: "+errorMsg)
}

func (as *authService) SignUp(ctx context.Context, request model.SignUpReq) (*model.SignUpRes, error) {
	reqID, _ := ctx.Value("request_id").(string)
	log := as.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   "Auth.SignUp",
		"email":      request.Email,
	})

	email := strings.ToLower(strings.TrimSpace(request.Email))

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

	// 3-layer rate limit check using atomic Lua script
	if err := as.checkRateLimit(ctx, email, as.ResendCooldown, as.MaxSignupPerHour); err != nil {
		log.WithError(err).Error("signup rate limit exceeded")
		return nil, err
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

	hashOTP := otp.Hash(codeOTP)

	pendingKey := "signup:pending:" + email
	otpKey := "signup:otp:" + email
	if err := as.Redis.Set(ctx, pendingKey, string(pendingBytes), as.OTPTtl).Err(); err != nil {
		log.WithError(err).Error("failed to store pending signup in redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store signup session")
	}
	if err := as.Redis.Set(ctx, otpKey, hashOTP, as.OTPTtl).Err(); err != nil {
		log.WithError(err).Error("failed to store otp in redis")
		_ = as.Redis.Del(ctx, pendingKey).Err()
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to store otp")
	}

	if err := as.Mail.SendTemplate(ctx, []string{email}, "GoCommerce OTP Verify", "otp_signup.html", mailer.OtpData{
		Code:    codeOTP,
		Expired: int(as.OTPTtl.Minutes()),
	}); err != nil {
		log.WithError(err).Error("failed to send otp email")
		_ = as.Redis.Del(ctx, pendingKey, otpKey).Err()
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to send otp to email")
	}

	res := model.SignUpRes{
		Email:     email,
		ExpiredAt: time.Now().Add(as.OTPTtl),
	}

	return &res, nil
}

func (as *authService) VerifyOTP(ctx context.Context, request model.VerifyOTP) (*model.VerifyOTPRes, error) {
	reqID, _ := ctx.Value("request_id").(string)
	log := as.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   "Auth.VerifyOTP",
		"email":      request.Email,
	})

	email := strings.ToLower(strings.TrimSpace(request.Email))

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	otpKey := "signup:otp:" + email
	pendingKey := "signup:pending:" + email

	storedHash, err := as.Redis.Get(ctx, otpKey).Result()
	if err == redis.Nil {
		log.Error("otp expired or not found")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "OTP expired or invalid")
	}
	if err != nil {
		log.WithError(err).Error("failed to get otp from redis")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to verify otp")
	}

	if !otp.Verify(request.CodeOTP, storedHash) {
		log.Error("invalid otp code")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid OTP code")
	}

	pendingData, err := as.Redis.Get(ctx, pendingKey).Result()
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

	exists, err := as.AuthRepo.Exists(ctx, "email", email)
	if err != nil {
		log.WithError(err).Error("failed to check user existence")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create user")
	}
	if exists {
		log.Error("email already registered")
		_ = as.Redis.Del(ctx, pendingKey, otpKey).Err()
		return nil, apperrors.Wrap(nil, http.StatusConflict, "email already registered")
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

	createdUser, err := as.AuthRepo.Create(ctx, user)
	if err != nil {
		log.WithError(err).Error("failed to create user")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create user")
	}

	tokenPair, err := jwt.GenerateTokenPair(
		createdUser.ID.String(),
		createdUser.Email,
		string(createdUser.Role),
		createdUser.IsVerified,
		as.JWTSecret,
		as.JWTAccessExpiry,
		as.JWTRefreshExpiry,
		as.JWTIssuer,
	)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to generate tokens")
	}

	_ = as.Redis.Del(ctx, pendingKey, otpKey).Err()
	_ = as.Redis.Del(ctx, "rl:rate:signup:minute:"+email).Err()
	_ = as.Redis.Del(ctx, "rl:rate:signup:hour:"+email).Err()
	_ = as.Redis.Del(ctx, "rl:rate:signup:day:"+email).Err()

	res := &model.VerifyOTPRes{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return res, nil
}

func (as *authService) ResendOTP(ctx context.Context, request model.ResendOTPReq) (*model.ResendOTPRes, error) {
	reqID, _ := ctx.Value("request_id").(string)
	log := as.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   "Auth.ResendOTP",
		"email":      request.Email,
	})

	email := strings.ToLower(strings.TrimSpace(request.Email))

	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	pendingKey := "signup:pending:" + email
	otpKey := "signup:otp:" + email

	exists, err := as.Redis.Exists(ctx, pendingKey).Result()
	if err != nil {
		log.WithError(err).Error("failed to check pending signup")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to check signup session")
	}
	if exists == 0 {
		log.Error("no pending signup found")
		return nil, apperrors.Wrap(nil, http.StatusBadRequest, "no pending signup found, please signup first")
	}

	// 3-layer rate limit check for resend using atomic Lua script
	if err := as.checkRateLimit(ctx, email, as.ResendCooldown, as.MaxSignupPerHour); err != nil {
		log.WithError(err).Error("resend rate limit exceeded")
		return nil, err
	}

	codeOTP, err := otp.Generate()
	if err != nil {
		log.WithError(err).Error("failed to generate otp")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed generated otp")
	}

	hashOTP := otp.Hash(codeOTP)

	if err := as.Redis.Set(ctx, otpKey, hashOTP, as.OTPTtl).Err(); err != nil {
		log.WithError(err).Error("failed to store otp in redis")
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
