package service

import (
	"database/sql"
	"errors"
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/auth"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type authServiceImpl struct {
	repo        repository.Store
	jwtSecret   string
	jwtDuration time.Duration
	log         *logrus.Logger
}

// NewAuthServiceImpl создает новый AuthService.
func NewAuthServiceImpl(repo repository.Store, jwtSecret string, jwtDuration time.Duration) AuthService {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return &authServiceImpl{
		repo:        repo,
		jwtSecret:   jwtSecret,
		jwtDuration: jwtDuration,
		log:         logger,
	}
}

func (s *authServiceImpl) Register(login, password string) (*domain.AuthResponse, *domain.AppError) {
	s.log.Infof("Register service called with login: %s", login)

	if login == "" || password == "" {
		s.log.Warnf("Registration failed for login '%s': Missing login or password", login)
		return nil, domain.ErrBadRequest("Login and password are required")
	}
	tempUser := domain.User{Login: login, Password: password}
	if err := tempUser.Validate(); err != nil {
		s.log.Warnf("Registration validation failed for login '%s': %v", login, err)
		return nil, domain.ErrValidationFailed(err)
	}

	s.log.Debugf("Checking if user '%s' exists in DB before registration", login)
	existingUser, findErr := s.repo.User().FindByLogin(login)

	if findErr == nil {
		s.log.Warnf("Registration attempt failed: Login '%s' already exists (User ID: %d)", login, existingUser.ID)
		return nil, domain.ErrDuplicateLogin(login)
	}

	var appErr *domain.AppError
	if errors.As(findErr, &appErr) {
		if appErr.HTTPStatus == http.StatusNotFound {
			s.log.Debugf("User with login '%s' not found, proceeding with registration.", login)
		} else {
			s.log.Errorf("Error checking existing user during registration for login %s: %v", login, findErr)
			return nil, appErr
		}
	} else {
		s.log.Errorf("Unexpected error type checking existing user for login %s: %v", login, findErr)
		return nil, domain.ErrInternalServer(fmt.Errorf("unexpected error checking user existence: %w", findErr))
	}

	s.log.Debugf("Hashing password for login '%s'", login)
	hashedPassword, hashErr := auth.HashPassword(password)
	if hashErr != nil {
		s.log.Errorf("Failed to hash password during registration for login %s: %v", login, hashErr)
		return nil, hashErr
	}
	user := &domain.User{
		Login:          login,
		HashedPassword: hashedPassword,
	}

	s.log.Debugf("Attempting to create user '%s' in DB", login)
	if createErr := s.repo.User().Create(user); createErr != nil {
		s.log.Errorf("Failed to create user in DB during registration for login %s: %v", login, createErr)
		if appErr, ok := createErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed creating user: %w", createErr))
	}
	s.log.Infof("User created successfully: ID=%d, Login=%s", user.ID, user.Login)

	s.log.Debugf("Attempting to create fridge for user ID %d", user.ID)
	fridge := &domain.Fridge{
		UserID: user.ID,
		Level:  1,
	}
	if createFridgeErr := s.repo.Fridge().Create(fridge); createFridgeErr != nil {
		s.log.Errorf("Failed to create fridge for user ID %d during registration: %v", user.ID, createFridgeErr)
		if appErr, ok := createFridgeErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed creating fridge: %w", createFridgeErr))
	}
	s.log.Infof("Fridge created successfully for user ID %d: FridgeID=%d, Code=%s", user.ID, fridge.ID, fridge.Code)

	s.log.Debugf("Attempting to link fridge code '%s' to user ID %d", fridge.Code, user.ID)
	user.FridgeCode = sql.NullString{String: fridge.Code, Valid: true}
	if updateErr := s.repo.User().Update(user); updateErr != nil {
		s.log.Errorf("Failed to link fridge code %s to user ID %d after registration: %v", fridge.Code, user.ID, updateErr)
	} else {
		s.log.Infof("Successfully linked fridge code '%s' to user ID %d", fridge.Code, user.ID)
	}

	s.log.Debugf("Generating JWT for user ID %d", user.ID)
	token, jwtErr := auth.GenerateJWT(user.ID, s.jwtSecret, s.jwtDuration)
	if jwtErr != nil {
		s.log.Errorf("Failed to generate JWT for user ID %d after registration: %v", user.ID, jwtErr)
		return nil, jwtErr
	}

	user.HashedPassword = ""
	response := &domain.AuthResponse{
		Token: token,
		User:  user,
	}
	s.log.Infof("Registration successful for user ID %d, Login %s", user.ID, user.Login)
	return response, nil
}

func (s *authServiceImpl) Login(login, password string) (*domain.AuthResponse, *domain.AppError) {
	s.log.Infof("Login service called for login: %s", login)

	if login == "" || password == "" {
		s.log.Warnf("Login failed for '%s': Missing login or password", login)
		return nil, domain.ErrBadRequest("Login and password are required")
	}

	s.log.Debugf("Attempting to find user by login '%s'", login)
	user, findErr := s.repo.User().FindByLogin(login)
	if findErr != nil {
		var appErr *domain.AppError
		if errors.As(findErr, &appErr) && appErr.HTTPStatus == http.StatusNotFound {
			s.log.Warnf("Login attempt failed for login '%s': User not found", login)
			return nil, domain.ErrUnauthorized()
		}
		s.log.Errorf("Error finding user during login for login %s: %v", login, findErr)
		if appErr != nil {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed finding user: %w", findErr))
	}

	s.log.Debugf("Checking password for user '%s' (ID: %d)", login, user.ID)
	if !auth.CheckPassword(user.HashedPassword, password) {
		s.log.Warnf("Login attempt failed for login '%s': Invalid password", login)
		return nil, domain.ErrUnauthorized()
	}

	s.log.Debugf("Generating JWT for user ID %d", user.ID)
	token, jwtErr := auth.GenerateJWT(user.ID, s.jwtSecret, s.jwtDuration)
	if jwtErr != nil {
		s.log.Errorf("Failed to generate JWT for user ID %d during login: %v", user.ID, jwtErr)
		return nil, jwtErr
	}

	user.HashedPassword = ""
	response := &domain.AuthResponse{
		Token: token,
		User:  user,
	}
	s.log.Infof("Login successful for user ID %d, Login %s", user.ID, user.Login)
	return response, nil
}
