package service

import (
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/auth"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type userServiceImpl struct {
	repo     repository.Store
	log      *logrus.Logger
	validate *validator.Validate
}

func NewUserServiceImpl(repo repository.Store) UserService {
	return &userServiceImpl{
		repo:     repo,
		log:      logrus.New(),
		validate: validator.New(),
	}
}

// GetUserProfile retrieves user details by ID.
func (s *userServiceImpl) GetUserProfile(userID int64) (*domain.User, *domain.AppError) {
	user, err := s.repo.User().FindByID(userID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); !ok || appErr.HTTPStatus != 404 {
			s.log.Errorf("Error fetching profile for user ID %d: %v", userID, err)
		}
		if appErr, ok := err.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed getting user profile: %w", err))
	}
	user.HashedPassword = ""
	return user, nil
}

// ChangePassword updates the user's password.
func (s *userServiceImpl) ChangePassword(userID int64, req domain.ChangePasswordRequest) *domain.AppError {
	err := s.validate.Var(req.NewPassword, "required,min=8,max=40")
	if err != nil {
		s.log.Warnf("Change password validation failed for user ID %d: %v", userID, err)
		return domain.ErrValidationFailed(err)
	}

	user, findErr := s.repo.User().FindByID(userID)
	if findErr != nil {
		s.log.Errorf("User not found during password change for user ID %d: %v", userID, findErr)
		if appErr, ok := findErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed finding user for password change: %w", findErr))
	}

	if !auth.CheckPassword(user.HashedPassword, req.CurrentPassword) {
		s.log.Warnf("Password change attempt failed for user ID %d: Invalid current password", userID)
		return domain.ErrUnauthorized()
	}

	newHashedPassword, hashErr := auth.HashPassword(req.NewPassword)
	if hashErr != nil {
		s.log.Errorf("Failed to hash new password for user ID %d: %v", userID, hashErr)
		return hashErr
	}
	user.HashedPassword = newHashedPassword

	if updateErr := s.repo.User().Update(user); updateErr != nil {
		s.log.Errorf("Failed to update user password in DB for user ID %d: %v", userID, updateErr)
		if appErr, ok := updateErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed updating password: %w", updateErr))
	}

	s.log.Infof("Password changed successfully for user ID %d", userID)
	return nil
}

func (s *userServiceImpl) DeleteUser(userID int64) *domain.AppError {
	err := s.repo.User().Delete(userID)
	if err != nil {
		s.log.Errorf("Failed to delete user ID %d: %v", userID, err)
		if appErr, ok := err.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed deleting user: %w", err))
	}
	s.log.Infof("User deleted successfully: ID=%d", userID)
	return nil
}
