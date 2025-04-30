package repository

import (
	"database/sql"
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlProductRepository struct {
	db *sqlx.DB
}

func newSQLProductRepository(db *sqlx.DB) ProductRepository {
	return &sqlProductRepository{db: db}
}

func (r *sqlProductRepository) Create(product *domain.Product) error {
	query := `INSERT INTO products (fridge_id, name, category, quantity, unit, expiry_date)
              VALUES (:fridge_id, :name, :category, :quantity, :unit, :expiry_date)
              RETURNING id`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return domain.ErrDatabase(fmt.Errorf("failed to prepare product create statement: %w", err))
	}
	defer stmt.Close()

	err = stmt.QueryRowx(product).Scan(&product.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			if pqErr.Constraint == "fk_fridge" { // Check if it's the fridge foreign key
				return domain.ErrNotFound("Fridge", product.FridgeID) // Assume fridge doesn't exist
			}
		}
		return domain.ErrDatabase(fmt.Errorf("failed to execute product create statement: %w", err))
	}
	return nil
}

// FindByID retrieves a product by its ID.
func (r *sqlProductRepository) FindByID(id int64) (*domain.Product, error) {
	product := &domain.Product{}
	query := `SELECT * FROM products WHERE id = $1`
	err := r.db.Get(product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("Product", id)
		}
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query product by id: %w", err))
	}
	return product, nil
}

// ListByFridgeID retrieves products for a given fridge with filtering and sorting.
func (r *sqlProductRepository) ListByFridgeID(params domain.ProductListParams) ([]domain.Product, error) {
	products := []domain.Product{}
	args := []interface{}{params.FridgeID}
	argID := 2

	query := `SELECT * FROM products WHERE fridge_id = $1`

	if params.Category != nil {
		query += fmt.Sprintf(" AND category = $%d", argID)
		args = append(args, *params.Category)
		argID++
	}
	if params.Search != nil && *params.Search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argID)
		args = append(args, "%"+*params.Search+"%")
		argID++
	}

	orderBy := "expiry_date ASC"
	if params.SortBy != "" {
		switch params.SortBy {
		case "expiry_desc":
			orderBy = "expiry_date DESC"
		case "name_asc":
			orderBy = "name ASC"
		case "name_desc":
			orderBy = "name DESC"
		}
	}
	query += " ORDER BY " + orderBy

	err := r.db.Select(&products, query, args...)
	if err != nil {
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query products by fridge_id: %w", err))
	}
	return products, nil
}

// Update modifies an existing product.
func (r *sqlProductRepository) Update(product *domain.Product) error {
	if product.ID == 0 {
		return domain.ErrBadRequest("Product ID is required for update")
	}
	query := `UPDATE products SET
                fridge_id = :fridge_id, name = :name, category = :category,
                quantity = :quantity, unit = :unit, expiry_date = :expiry_date
              WHERE id = :id`

	res, err := r.db.NamedExec(query, product)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			if pqErr.Constraint == "fk_fridge" {
				return domain.ErrNotFound("Fridge", product.FridgeID)
			}
		}
		return domain.ErrDatabase(fmt.Errorf("failed to execute product update statement: %w", err))
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return domain.ErrNotFound("Product", product.ID)
	}
	return nil
}

// Delete removes a product from the database.
func (r *sqlProductRepository) Delete(id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return domain.ErrDatabase(fmt.Errorf("failed to execute product delete statement: %w", err))
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return domain.ErrNotFound("Product", id)
	}
	return nil
}

// UpdateQuantity increments or decrements the quantity of a product.
func (r *sqlProductRepository) UpdateQuantity(productID int64, change float64) (*domain.Product, error) {
	product := &domain.Product{}
	query := `UPDATE products
              SET quantity = quantity + $1
              WHERE id = $2 AND quantity + $1 >= 0 
              RETURNING *`

	err := r.db.Get(product, query, change, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			existingProd, findErr := r.FindByID(productID)
			if findErr != nil {
				if appErr, ok := findErr.(*domain.AppError); ok && appErr.HTTPStatus == 404 {
					return nil, domain.ErrNotFound("Product", productID)
				}
				return nil, findErr
			}
			if existingProd.Quantity+change < 0 {
				return nil, domain.ErrBadRequest(fmt.Sprintf("Cannot decrease quantity by %.2f, current quantity is %.2f", -change, existingProd.Quantity))
			}
			return nil, domain.ErrDatabase(fmt.Errorf("update quantity failed for unknown reason for product %d", productID))
		}
		return nil, domain.ErrDatabase(fmt.Errorf("failed to update product quantity: %w", err))
	}
	return product, nil
}

// CountByFridgeID returns the total number of products in a fridge.
func (r *sqlProductRepository) CountByFridgeID(fridgeID int64) (int64, error) {
	var count int64
	query := `SELECT count(*) FROM products WHERE fridge_id = $1`
	err := r.db.Get(&count, query, fridgeID)
	if err != nil {
		return 0, domain.ErrDatabase(fmt.Errorf("failed to count products by fridge_id: %w", err))
	}
	return count, nil
}

// GetProductsExpiringOnDate retrieves products expiring on a specific date (midnight to midnight).
func (r *sqlProductRepository) GetProductsExpiringOnDate(fridgeID int64, date int64) ([]domain.Product, error) {
	products := []domain.Product{}
	startOfDay := date
	endOfDay := startOfDay + (24 * 60 * 60) - 1

	query := `SELECT * FROM products WHERE fridge_id = $1 AND expiry_date >= $2 AND expiry_date <= $3 ORDER BY name ASC`
	err := r.db.Select(&products, query, fridgeID, startOfDay, endOfDay)
	if err != nil {
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query products expiring on date: %w", err))
	}
	return products, nil
}

// GetProductsExpiringSoon retrieves products expiring within the next N days for notifications.
func (r *sqlProductRepository) GetProductsExpiringSoon(fridgeID int64, notificationDays []int) ([]domain.Product, error) {
	if len(notificationDays) == 0 {
		return []domain.Product{}, nil // Nothing to check
	}

	products := []domain.Product{}
	now := time.Now().Unix()
	placeholders := []string{}
	args := []interface{}{fridgeID, now}
	argID := 3

	for _, days := range notificationDays {
		targetDayStart := time.Now().AddDate(0, 0, days).Truncate(24 * time.Hour).Unix()
		targetDayEnd := targetDayStart + (24 * 60 * 60) - 1
		placeholders = append(placeholders, fmt.Sprintf("(expiry_date >= $%d AND expiry_date <= $%d)", argID, argID+1))
		args = append(args, targetDayStart, targetDayEnd)
		argID += 2
	}

	expiryCondition := strings.Join(placeholders, " OR ")
	query := fmt.Sprintf(`SELECT * FROM products WHERE fridge_id = $1 AND expiry_date >= $2 AND (%s) ORDER BY expiry_date ASC`, expiryCondition)

	err := r.db.Select(&products, query, args...)
	if err != nil {
		// sql.ErrNoRows is not an error for Select
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query products expiring soon: %w", err))
	}
	return products, nil
}
