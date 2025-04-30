package service

import (
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"sync"

	"github.com/sirupsen/logrus"
)

type fridgeServiceImpl struct {
	repo             repository.Store
	log              *logrus.Logger
	achievementMutex sync.Mutex
}

func NewFridgeServiceImpl(repo repository.Store) FridgeService {
	return &fridgeServiceImpl{
		repo: repo,
		log:  logrus.New(),
	}
}

// GetFridgeByUserID fetches the fridge belonging to the user.
func (s *fridgeServiceImpl) GetFridgeByUserID(userID int64) (*domain.Fridge, *domain.AppError) {
	fridge, err := s.repo.Fridge().FindByUserID(userID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); !ok || appErr.HTTPStatus != 404 {
			s.log.Errorf("Error fetching fridge for user ID %d: %v", userID, err)
		}
		if appErr, ok := err.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed getting fridge by user ID: %w", err))
	}
	return fridge, nil
}

// UpdateNotificationSettings updates the fridge's notification preferences.
func (s *fridgeServiceImpl) UpdateNotificationSettings(userID int64, req domain.UpdateNotificationSettingsRequest) *domain.AppError {
	fridge, err := s.GetFridgeByUserID(userID)
	if err != nil {
		return err
	}

	updateErr := s.repo.Fridge().UpdateNotificationSettings(fridge.ID, &req)
	if updateErr != nil {
		s.log.Errorf("Failed to update notification settings for fridge ID %d (User ID %d): %v", fridge.ID, userID, updateErr)
		if appErr, ok := updateErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed updating notification settings: %w", updateErr))
	}

	s.log.Infof("Notification settings updated for fridge ID %d (User ID %d)", fridge.ID, userID)
	return nil
}

// GetFridgeInfo combines fridge data with calculated level info.
func (s *fridgeServiceImpl) GetFridgeInfo(userID int64) (*FridgeInfoResponse, *domain.AppError) {
	fridge, err := s.GetFridgeByUserID(userID)
	if err != nil {
		return nil, err
	}

	currentLevel := getLevel(fridge.ConsumedAll)
	needsDBUpdate := false

	if fridge.Level != currentLevel {
		s.log.Infof("Fridge level discrepancy detected for Fridge ID %d (User ID %d). DB Level: %d, Calculated Level: %d. Updating DB.", fridge.ID, userID, fridge.Level, currentLevel)
		fridge.Level = currentLevel
		needsDBUpdate = true
	}

	if currentLevel >= 2 && !fridge.AchEquator {
		granted, achErr := s.CheckAndGrantAchievement(fridge.ID, AchievementEquator, nil)
		if achErr != nil {
			s.log.Errorf("Error checking/granting Equator achievement for Fridge ID %d: %v", fridge.ID, achErr)
		} else if granted {
			s.log.Infof("Equator achievement granted for Fridge ID %d", fridge.ID)
			fridge.AchEquator = true
		}
	}

	if needsDBUpdate {
		if updateErr := s.repo.Fridge().Update(fridge); updateErr != nil {
			s.log.Errorf("Failed to update fridge level in DB for Fridge ID %d: %v", fridge.ID, updateErr)
		}
	}

	productsToNext := getProductsToNextLevel(fridge.ConsumedAll)
	description := getFridgeDescription(currentLevel)

	response := &FridgeInfoResponse{
		Fridge:              *fridge,
		ProductsToNextLevel: productsToNext,
		Description:         description,
	}

	return response, nil
}

// CheckAndGrantAchievement checks if an achievement is already granted and grants it if not.
func (s *fridgeServiceImpl) CheckAndGrantAchievement(fridgeID int64, achievementType string, relatedData interface{}) (bool, *domain.AppError) {
	s.achievementMutex.Lock()
	defer s.achievementMutex.Unlock()

	fridge, err := s.repo.Fridge().FindByID(fridgeID)
	if err != nil {
		s.log.Errorf("Failed to get fridge %d for achievement check (%s): %v", fridgeID, achievementType, err)
		if appErr, ok := err.(*domain.AppError); ok {
			return false, appErr
		}
		return false, domain.ErrInternalServer(fmt.Errorf("failed finding fridge for achievement check: %w", err))
	}

	var needsUpdate bool
	var hello, equator, salad *bool

	switch achievementType {
	case AchievementHelloWorld:
		if !fridge.AchHelloWorld {
			s.log.Infof("Condition met for achievement '%s' for fridge %d.", achievementType, fridgeID)
			needsUpdate = true
			flag := true
			hello = &flag
		}
	case AchievementEquator:
		if !fridge.AchEquator && fridge.Level >= 2 {
			s.log.Infof("Condition met for achievement '%s' for fridge %d.", achievementType, fridgeID)
			needsUpdate = true
			flag := true
			equator = &flag
		}
	case AchievementSalad:
		category, ok := relatedData.(domain.ProductCategory)
		if !fridge.AchSalad && ok && category == domain.CategoryVegetables {
			s.log.Infof("Condition met for achievement '%s' for fridge %d.", achievementType, fridgeID)
			needsUpdate = true
			flag := true
			salad = &flag
		}
	default:
		s.log.Warnf("Unknown achievement type checked for fridge %d: %s", fridgeID, achievementType)
		return false, domain.ErrBadRequest(fmt.Sprintf("Unknown achievement type: %s", achievementType))
	}

	if needsUpdate {
		updateErr := s.repo.Fridge().UpdateAchievements(fridgeID, hello, equator, salad)
		if updateErr != nil {
			s.log.Errorf("Failed to grant achievement '%s' for fridge %d: %v", achievementType, fridgeID, updateErr)
			if appErr, ok := updateErr.(*domain.AppError); ok {
				return false, appErr
			}
			return false, domain.ErrInternalServer(fmt.Errorf("failed updating achievements: %w", updateErr))
		}
		s.log.Infof("Achievement '%s' granted and updated in DB for fridge %d", achievementType, fridgeID)
		return true, nil
	}

	return false, nil
}
