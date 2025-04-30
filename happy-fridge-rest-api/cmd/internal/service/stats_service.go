package service

import (
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"sort"

	"github.com/sirupsen/logrus"
)

type statsServiceImpl struct {
	repo repository.Store
	log  *logrus.Logger
}

func NewStatsServiceImpl(repo repository.Store) StatsService {
	return &statsServiceImpl{
		repo: repo,
		log:  logrus.New(),
	}
}

func (s *statsServiceImpl) GetFridgeStatistics(userID int64) (*domain.FridgeStats, *domain.AppError) {

	fridge, fridgeErr := s.repo.Fridge().FindByUserID(userID)
	if fridgeErr != nil {
		if appErr, ok := fridgeErr.(*domain.AppError); ok && appErr.HTTPStatus == 404 {
			s.log.Warnf("Fridge not found for user %d when calculating stats.", userID)
			return nil, domain.ErrFridgeNotFoundForUser(userID)
		}
		s.log.Errorf("Error finding fridge for user %d for stats: %v", userID, fridgeErr)
		if appErr, ok := fridgeErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed finding fridge for stats: %w", fridgeErr))
	}

	consumedOnTimePercentage := 0.0
	if fridge.ConsumedAll > 0 {
		consumedOnTimePercentage = (float64(fridge.ConsumedOnTime) / float64(fridge.ConsumedAll)) * 100.0
	}

	expiredCounts := map[string]int64{
		string(domain.CategoryMeat):       fridge.ConsumedMeatExp,
		string(domain.CategoryMilk):       fridge.ConsumedMilkExp,
		string(domain.CategoryVegetables): fridge.ConsumedVegetablesExp,
		string(domain.CategoryFruit):      fridge.ConsumedFruitExp,
		string(domain.CategorySweets):     fridge.ConsumedSweetsExp,
		string(domain.CategoryBakery):     fridge.ConsumedBakeryExp,
		string(domain.CategorySeafood):    fridge.ConsumedSeafoodExp,
		string(domain.CategoryCooked):     fridge.ConsumedCookedExp,
		string(domain.CategoryOther):      fridge.ConsumedOtherExp,
	}

	totalExpired := int64(0)
	for _, count := range expiredCounts {
		totalExpired += count
	}

	expiredCategoryPercentage := make(map[string]float64)
	otherPercentage := 0.0
	otherCount := int64(0)

	if totalExpired > 0 {
		type categoryCount struct {
			Name  string
			Count int64
		}
		var sortedCategories []categoryCount
		for name, count := range expiredCounts {
			if count > 0 {
				sortedCategories = append(sortedCategories, categoryCount{Name: name, Count: count})
			}
		}

		sort.Slice(sortedCategories, func(i, j int) bool {
			return sortedCategories[i].Count > sortedCategories[j].Count
		})

		for i, cat := range sortedCategories {
			percentage := (float64(cat.Count) / float64(totalExpired)) * 100.0
			if i < 4 {
				expiredCategoryPercentage[cat.Name] = percentage
			} else {
				otherPercentage += percentage
				otherCount += cat.Count
			}
		}

		if otherCount > 0 {
			expiredCategoryPercentage[string(domain.CategoryOther)] = otherPercentage
		}

	}

	stats := &domain.FridgeStats{
		ConsumedOnTimePercentage:  consumedOnTimePercentage,
		ExpiredByCategory:         expiredCounts,
		ExpiredCategoryPercentage: expiredCategoryPercentage,
	}

	return stats, nil
}
