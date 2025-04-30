package service

import (
	"errors"
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type productServiceImpl struct {
	repo          repository.Store
	fridgeService FridgeService
	log           *logrus.Logger
	validate      *validator.Validate
}

// NewProductServiceImpl creates a new ProductService.
func NewProductServiceImpl(repo repository.Store, fridgeService FridgeService) ProductService {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return &productServiceImpl{
		repo:          repo,
		fridgeService: fridgeService,
		log:           logger,
		validate:      validator.New(),
	}
}

func (s *productServiceImpl) getFridgeAndCheckOwnership(userID int64) (*domain.Fridge, *domain.AppError) {
	fridge, err := s.fridgeService.GetFridgeByUserID(userID)
	if err != nil {
		return nil, err
	}
	return fridge, nil
}

func (s *productServiceImpl) getProductAndCheckOwnership(userID, productID int64) (*domain.Product, *domain.AppError) {
	product, err := s.repo.Product().FindByID(productID)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}

		s.log.Errorf("Unexpected error type getting product %d: %v", productID, err)
		return nil, domain.ErrInternalServer(fmt.Errorf("failed getting product: %w", err))
	}

	fridge, fridgeErr := s.getFridgeAndCheckOwnership(userID)
	if fridgeErr != nil {
		return nil, fridgeErr
	}
	if product.FridgeID != fridge.ID {
		s.log.Warnf("User %d attempted to access product %d belonging to fridge %d (their fridge is %d)", userID, productID, product.FridgeID, fridge.ID)
		return nil, domain.ErrForbidden()
	}

	product.CalculateDaysLeft()
	return product, nil
}

// AddProduct adds a new product to the user's fridge.
func (s *productServiceImpl) AddProduct(userID int64, req domain.ProductCreateRequest) (*domain.Product, *domain.AppError) {
	s.log.Debugf("AddProduct called for user %d with data: %+v", userID, req)
	if err := s.validate.Struct(req); err != nil {
		s.log.Warnf("Add product validation failed for user %d: %v", userID, err)
		return nil, domain.ErrValidationFailed(err)
	}

	fridge, fridgeErr := s.getFridgeAndCheckOwnership(userID)
	if fridgeErr != nil {
		return nil, fridgeErr
	}

	expiryTime, parseErr := time.Parse("2006-01-02", req.ExpiryDate)
	if parseErr != nil {
		s.log.Warnf("Invalid expiry date format '%s' for user %d: %v", req.ExpiryDate, userID, parseErr)
		return nil, domain.ErrBadRequest(fmt.Sprintf("Invalid expiry date format. Use YYYY-MM-DD: %v", parseErr))
	}
	expiryTimestamp := expiryTime.Unix()

	product := &domain.Product{
		FridgeID:   fridge.ID,
		Name:       req.Name,
		Category:   req.Category,
		Quantity:   req.Quantity,
		Unit:       req.Unit,
		ExpiryDate: expiryTimestamp,
	}

	if createErr := s.repo.Product().Create(product); createErr != nil {
		s.log.Errorf("Failed to create product in DB for fridge %d (User %d): %v", fridge.ID, userID, createErr)
		if appErr, ok := createErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed creating product: %w", createErr))
	}
	s.log.Infof("Product created successfully (ID: %d) for fridge %d (User %d)", product.ID, fridge.ID, userID)

	granted, achErr := s.fridgeService.CheckAndGrantAchievement(fridge.ID, AchievementHelloWorld, nil)
	if achErr != nil {
		s.log.Errorf("Error checking/granting HelloWorld achievement for Fridge ID %d after adding product: %v", fridge.ID, achErr)
	} else if granted {
		s.log.Infof("HelloWorld achievement granted for Fridge ID %d", fridge.ID)
	}
	product.CalculateDaysLeft()

	return product, nil
}

// ListProducts retrieves products for the user's fridge.
func (s *productServiceImpl) ListProducts(userID int64, params domain.ProductListParams) ([]domain.Product, *domain.AppError) {
	s.log.Debugf("ListProducts called for user %d with params: %+v", userID, params)
	fridge, fridgeErr := s.getFridgeAndCheckOwnership(userID)
	if fridgeErr != nil {
		return nil, fridgeErr
	}

	params.FridgeID = fridge.ID
	if params.SortBy == "" {
		params.SortBy = "expiry_asc"
	}
	products, listErr := s.repo.Product().ListByFridgeID(params)
	if listErr != nil {
		s.log.Errorf("Failed to list products for fridge %d (User %d): %v", fridge.ID, userID, listErr)
		if appErr, ok := listErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed listing products: %w", listErr))
	}

	for i := range products {
		products[i].CalculateDaysLeft()
	}
	s.log.Debugf("Found %d products for user %d", len(products), userID)
	return products, nil
}

// GetProductByID retrieves a specific product, checking ownership.
func (s *productServiceImpl) GetProductByID(userID, productID int64) (*domain.Product, *domain.AppError) {
	s.log.Debugf("GetProductByID called for user %d, product %d", userID, productID)
	product, err := s.getProductAndCheckOwnership(userID, productID)
	if err != nil {
		return nil, err
	}
	s.log.Debugf("Product %d found for user %d", productID, userID)
	return product, nil
}

// UpdateProduct updates product details.
func (s *productServiceImpl) UpdateProduct(userID, productID int64, req domain.ProductUpdateRequest) (*domain.Product, *domain.AppError) {
	s.log.Debugf("UpdateProduct called for user %d, product %d with data: %+v", userID, productID, req)
	// 1. Validate request payload (partial validation)
	if err := s.validate.Struct(req); err != nil {
		s.log.Warnf("Update product validation failed for user %d, product %d: %v", userID, productID, err)
		return nil, domain.ErrValidationFailed(err)
	}

	product, err := s.getProductAndCheckOwnership(userID, productID)
	if err != nil {
		return nil, err
	}

	updated := false
	if req.Name != nil {
		product.Name = *req.Name
		updated = true
	}
	if req.Category != nil {
		product.Category = *req.Category
		updated = true
	}
	if req.Quantity != nil {
		if *req.Quantity < 0 {
			s.log.Warnf("Update product failed for user %d, product %d: Negative quantity provided (%.2f)", userID, productID, *req.Quantity)
			return nil, domain.ErrBadRequest("Quantity cannot be negative")
		}
		product.Quantity = *req.Quantity
		updated = true
	}
	if req.Unit != nil {
		product.Unit = *req.Unit
		updated = true
	}
	if req.ExpiryDate != nil {
		expiryTime, parseErr := time.Parse("2006-01-02", *req.ExpiryDate)
		if parseErr != nil {
			s.log.Warnf("Invalid expiry date format '%s' during update for user %d, product %d: %v", *req.ExpiryDate, userID, productID, parseErr)
			return nil, domain.ErrBadRequest("Invalid expiry date format. Use YYYY-MM-DD.")
		}
		product.ExpiryDate = expiryTime.Unix()
		updated = true
	}

	if updated {
		s.log.Debugf("Attempting to update product %d in DB for user %d", productID, userID)
		if updateErr := s.repo.Product().Update(product); updateErr != nil {
			s.log.Errorf("Failed to update product %d in DB (User %d): %v", productID, userID, updateErr)
			if appErr, ok := updateErr.(*domain.AppError); ok {
				return nil, appErr
			}
			return nil, domain.ErrInternalServer(fmt.Errorf("failed updating product: %w", updateErr))
		}
		s.log.Infof("Product %d updated successfully (User %d)", productID, userID)
	} else {
		s.log.Infof("Product %d update requested by user %d, but no changes provided.", productID, userID)
	}

	product.CalculateDaysLeft()
	return product, nil
}

// DeleteProduct removes a product.
func (s *productServiceImpl) DeleteProduct(userID, productID int64) *domain.AppError {
	s.log.Debugf("DeleteProduct called for user %d, product %d", userID, productID)
	_, err := s.getProductAndCheckOwnership(userID, productID)
	if err != nil {
		return err
	}

	s.log.Debugf("Attempting to delete product %d from DB for user %d", productID, userID)
	deleteErr := s.repo.Product().Delete(productID)
	if deleteErr != nil {
		s.log.Errorf("Failed to delete product %d (User %d): %v", productID, userID, deleteErr)
		if appErr, ok := deleteErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed deleting product: %w", deleteErr))
	}
	s.log.Infof("Product %d deleted successfully (User %d)", productID, userID)
	return nil
}

// UpdateProductQuantity handles +/- actions.
func (s *productServiceImpl) UpdateProductQuantity(userID, productID int64, change float64) (*domain.Product, *domain.AppError) {
	s.log.Debugf("UpdateProductQuantity called for user %d, product %d, change signal: %.0f", userID, productID, change)
	product, err := s.getProductAndCheckOwnership(userID, productID)
	if err != nil {
		return nil, err
	}

	actualChange := 0.0
	if change == 1 { // Increment
		if product.Unit.IsCountable() {
			actualChange = 1.0
		} else if product.Unit.IsWeightOrVolume() {
			actualChange = 0.1
		} else {
			s.log.Warnf("Increment quantity requested for product %d with unknown unit type '%s'", productID, product.Unit)
			return nil, domain.ErrBadRequest("Cannot increment quantity for this unit type")
		}
	} else if change == -1 { // Decrement
		if product.Unit.IsCountable() {
			actualChange = -1.0
		} else if product.Unit.IsWeightOrVolume() {
			actualChange = -0.1
		} else {
			s.log.Warnf("Decrement quantity requested for product %d with unknown unit type '%s'", productID, product.Unit)
			return nil, domain.ErrBadRequest("Cannot decrement quantity for this unit type")
		}
	} else {
		s.log.Warnf("Invalid change value for quantity update: %.2f", change)
		return nil, domain.ErrBadRequest("Invalid change value for quantity update, expected 1 or -1")
	}

	if product.Quantity+actualChange < 0 {
		s.log.Warnf("Attempt to set negative quantity for product %d (User %d). Current: %.2f, Change: %.2f",
			productID, userID, product.Quantity, actualChange)
		return nil, domain.ErrBadRequest(fmt.Sprintf("Quantity cannot be reduced further (current: %.2f %s)", product.Quantity, product.Unit))
	}

	s.log.Debugf("Attempting to update quantity for product %d by %.2f", productID, actualChange)
	updatedProduct, updateErr := s.repo.Product().UpdateQuantity(productID, actualChange)
	if updateErr != nil {
		s.log.Errorf("Failed to update quantity for product %d (User %d): %v", productID, userID, updateErr)
		if appErr, ok := updateErr.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed updating quantity: %w", updateErr))
	}

	updatedProduct.CalculateDaysLeft()
	s.log.Infof("Quantity updated for product %d (User %d). New quantity: %.2f %s", productID, userID, updatedProduct.Quantity, updatedProduct.Unit)
	return updatedProduct, nil
}

func (s *productServiceImpl) ConsumeProduct(userID, productID int64) *domain.AppError {
	s.log.Debugf("ConsumeProduct called for user %d, product %d", userID, productID)
	product, err := s.getProductAndCheckOwnership(userID, productID)
	if err != nil {
		return err
	}

	fridgeID := product.FridgeID
	category := product.Category
	daysLeft := product.DaysLeft

	s.log.Debugf("Attempting to delete consumed product %d (DaysLeft: %d)", productID, daysLeft)
	deleteErr := s.repo.Product().Delete(productID)
	if deleteErr != nil {
		s.log.Errorf("Failed to delete consumed product %d (User %d): %v", productID, userID, deleteErr)
		if appErr, ok := deleteErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed deleting consumed product: %w", deleteErr))
	}
	s.log.Infof("Successfully deleted consumed product %d", productID)

	s.log.Debugf("Incrementing consumed_all for fridge %d", fridgeID)
	if incErr := s.repo.Fridge().IncrementCounter(fridgeID, "consumed_all", 1); incErr != nil {
		s.log.Errorf("Failed to increment consumed_all for fridge %d after consuming product %d: %v", fridgeID, productID, incErr)
		if appErr, ok := incErr.(*domain.AppError); ok {
			return appErr
		}
		return domain.ErrInternalServer(fmt.Errorf("failed incrementing consumed_all stat: %w", incErr))
	}
	s.log.Debugf("Successfully incremented consumed_all for fridge %d", fridgeID)

	if daysLeft >= 0 {
		s.log.Debugf("Product %d consumed on time (or today). Incrementing consumed_on_time for fridge %d", productID, fridgeID)
		if incErr := s.repo.Fridge().IncrementCounter(fridgeID, "consumed_on_time", 1); incErr != nil {
			s.log.Errorf("Failed to increment consumed_on_time for fridge %d after consuming product %d: %v", fridgeID, productID, incErr)
			if appErr, ok := incErr.(*domain.AppError); ok {
				return appErr
			}
			return domain.ErrInternalServer(fmt.Errorf("failed incrementing consumed_on_time stat: %w", incErr))
		}
		s.log.Debugf("Successfully incremented consumed_on_time for fridge %d", fridgeID)

		if category == domain.CategoryVegetables {
			granted, achErr := s.fridgeService.CheckAndGrantAchievement(fridgeID, AchievementSalad, category)
			if achErr != nil {
				s.log.Errorf("Error checking/granting Salad achievement for Fridge ID %d: %v", fridgeID, achErr)
			} else if granted {
				s.log.Infof("Salad achievement granted for Fridge ID %d", fridgeID)
			}
		}
	} else {
		expCounterField := category.GetConsumedExpFridgeField()
		s.log.Debugf("Product %d was expired (DaysLeft: %d). Incrementing %s for fridge %d", productID, daysLeft, expCounterField, fridgeID)
		if incErr := s.repo.Fridge().IncrementCounter(fridgeID, expCounterField, 1); incErr != nil {
			s.log.Errorf("Failed to increment %s for fridge %d after consuming expired product %d: %v", expCounterField, fridgeID, productID, incErr)
			if appErr, ok := incErr.(*domain.AppError); ok {
				return appErr
			}
			return domain.ErrInternalServer(fmt.Errorf("failed incrementing expired stat %s: %w", expCounterField, incErr))
		}
		s.log.Debugf("Successfully incremented %s for fridge %d", expCounterField, fridgeID)
	}

	s.log.Infof("Product %d consumed successfully (Expired: %t) by user %d from fridge %d. Stats updated.", productID, daysLeft < 0, userID, fridgeID)
	return nil
}

// GetProductsExpiringOnDate retrieves products expiring on a specific calendar date.
func (s *productServiceImpl) GetProductsExpiringOnDate(userID int64, date string) ([]domain.Product, *domain.AppError) {
	s.log.Debugf("GetProductsExpiringOnDate called for user %d, date %s", userID, date)
	fridge, fridgeErr := s.getFridgeAndCheckOwnership(userID)
	if fridgeErr != nil {
		return nil, fridgeErr
	}

	targetDate, parseErr := time.Parse("2006-01-02", date)
	if parseErr != nil {
		s.log.Warnf("Invalid date format '%s' for calendar view (User %d): %v", date, userID, parseErr)
		return nil, domain.ErrBadRequest("Invalid date format. Use YYYY-MM-DD.")
	}
	targetTimestamp := targetDate.Unix()

	s.log.Debugf("Querying products expiring on %s (timestamp %d) for fridge %d", date, targetTimestamp, fridge.ID)
	products, err := s.repo.Product().GetProductsExpiringOnDate(fridge.ID, targetTimestamp)
	if err != nil {
		s.log.Errorf("Failed to get products expiring on %s for fridge %d (User %d): %v", date, fridge.ID, userID, err)
		if appErr, ok := err.(*domain.AppError); ok {
			return nil, appErr
		}
		return nil, domain.ErrInternalServer(fmt.Errorf("failed getting calendar products: %w", err))
	}

	for i := range products {
		products[i].CalculateDaysLeft()
	}
	s.log.Debugf("Found %d products expiring on %s for user %d", len(products), date, userID)
	return products, nil
}
