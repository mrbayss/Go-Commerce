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
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/otp"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepo         repository.AuthRepository
	Log              *logrus.Logger
	Redis            *redis.Client
	Validation       *validator.Validate
	Mail             *mailer.Client
	OTPTtl           time.Duration
	MaxSignupEmail   int
	SignupRateWindow time.Duration
	ResendCooldown   time.Duration
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
	maxsignupemail int,
	signupratewindow time.Duration,
	resendcooldown time.Duration,
) *AuthService {
	if otpttl <= 0 {
		otpttl = 10 * time.Minute
	}
	if maxsignupemail <= 0 {
		maxsignupemail = 3
	}
	if signupratewindow <= 0 {
		signupratewindow = 10 * time.Minute
	}
	if resendcooldown <= 0 {
		resendcooldown = 60 * time.Second
	}
	return &AuthService{
		AuthRepo:         authRepo,
		Log:              log,
		Redis:            redis,
		Validation:       validation,
		Mail:             mail,
		OTPTtl:           otpttl,
		MaxSignupEmail:   maxsignupemail,
		SignupRateWindow: signupratewindow,
		ResendCooldown:   resendcooldown,
	}
}

func (as *AuthService) SignUp(ctx context.Context, request model.SignUpReq) (*model.SignUpRes, error) {
	// Extract request_id from context
	reqID, _ := ctx.Value("request_id").(string)
	log := as.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   "Auth.SignUp",
		"email":      request.Email,
	})

	email := strings.ToLower(strings.TrimSpace(request.Email))

	// validate struct
	if err := as.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	// Precheck duplicate account at DB (cheap, before paying bcrypt cost).
	exists, err := as.AuthRepo.Exists(ctx, "email", email)
	if err != nil {
		log.WithError(err).Error("failed to lookup user")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to lookup user")
	}
	if exists {
		log.Error("email already registered")
		return nil, apperrors.Wrap(nil, http.StatusConflict, "email already registered")
	}

	// Redis rate limit: max N signup attempts per email per window (anti spam).
	rateKey := "rl:signup:email:" + email
	n, err := as.Redis.Incr(ctx, rateKey).Result()
	if err != nil {
		log.WithError(err).Error("failed to rate limit signup")
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to rate limit signup")
	}
	if n == 1 {
		_ = as.Redis.Expire(ctx, rateKey, as.SignupRateWindow).Err()
	}
	if int(n) > as.MaxSignupEmail {
		log.WithField("attempts", n).Error("signup rate limit exceeded")
		return nil, apperrors.Wrap(nil, http.StatusTooManyRequests, "too many signup attempts, try again later")
	}

	// Hash password
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

	// Store pending signup + hashed OTP in Redis; account row is only created after OTP verification.
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
		// Roll back redis writes on email failure.
		_ = as.Redis.Del(ctx, pendingKey, otpKey).Err()
		return nil, apperrors.Wrap(err, http.StatusBadGateway, "failed to send otp to email")
	}

	res := model.SignUpRes{
		Email:     email,
		ExpiredAt: time.Now().Add(as.OTPTtl),
	}

	return &res, nil
}
