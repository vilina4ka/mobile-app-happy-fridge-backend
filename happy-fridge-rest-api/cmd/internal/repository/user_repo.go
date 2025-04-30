package repository

import (
	"database/sql"
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type sqlUserRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func newSQLUserRepository(db *sqlx.DB) UserRepository {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	return &sqlUserRepository{db: db, log: logger}
}

// Create inserts a new user into the database.
func (r *sqlUserRepository) Create(user *domain.User) error {
	r.log.Debugf("Attempting to create user with login: %s", user.Login)
	query := `INSERT INTO users (login, hashed_password, fridge_code)
              VALUES (:login, :hashed_password, :fridge_code)
              RETURNING id`

	params := map[string]interface{}{
		"login":           user.Login,
		"hashed_password": user.HashedPassword,
		"fridge_code":     user.FridgeCode,
	}

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		r.log.Errorf("Failed to prepare user create statement for login '%s': %v", user.Login, err)
		return domain.ErrDatabase(fmt.Errorf("failed to prepare user create statement: %w", err))
	}
	defer stmt.Close()

	err = stmt.QueryRowx(params).Scan(&user.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.log.Warnf("Failed to create user: Login '%s' already exists (DB constraint violation).", user.Login)
			return domain.ErrDuplicateLogin(user.Login) // Return specific 409 error
		}
		r.log.Errorf("Failed to execute user create statement for login '%s': %v", user.Login, err)
		return domain.ErrDatabase(fmt.Errorf("failed to execute user create statement: %w", err))
	}
	r.log.Infof("Successfully created user with login '%s', assigned ID: %d", user.Login, user.ID)
	return nil
}

// FindByLogin retrieves a user by their login.
func (r *sqlUserRepository) FindByLogin(login string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, login, hashed_password, fridge_code FROM users WHERE login = $1`
	r.log.Debugf("Executing FindByLogin query for login: %s", login)

	err := r.db.Get(user, query, login)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Debugf("User with login '%s' not found (sql.ErrNoRows)", login)
			return nil, domain.ErrUserNotFoundByLogin(login)
		}
		r.log.Errorf("Database error during FindByLogin for login '%s': %v", login, err)
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query user by login: %w", err))
	}

	r.log.Debugf("User found with login '%s': ID=%d", login, user.ID)
	return user, nil
}

// FindByID retrieves a user by their ID.
func (r *sqlUserRepository) FindByID(id int64) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, login, hashed_password, fridge_code FROM users WHERE id = $1`
	r.log.Debugf("Executing FindByID query for ID: %d", id)
	err := r.db.Get(user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warnf("User with ID %d not found (sql.ErrNoRows)", id)
			return nil, domain.ErrNotFound("User", id)
		}
		r.log.Errorf("Database error during FindByID for ID %d: %v", id, err)
		return nil, domain.ErrDatabase(fmt.Errorf("failed to query user by id: %w", err))
	}
	r.log.Debugf("User found with ID %d", id)
	return user, nil
}

// Update modifies an existing user (e.g., password update or linking fridge code).
func (r *sqlUserRepository) Update(user *domain.User) error {
	if user.ID == 0 {
		r.log.Error("Attempted to update user with ID 0")
		return domain.ErrBadRequest("User ID is required for update")
	}
	r.log.Debugf("Attempting to update user with ID: %d, Login: %s", user.ID, user.Login)

	query := `UPDATE users SET
                hashed_password = :hashed_password,
                fridge_code = :fridge_code,
                login = :login
              WHERE id = :id`

	res, err := r.db.NamedExec(query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.log.Warnf("Failed to update user %d: Login '%s' already exists.", user.ID, user.Login)
			return domain.ErrDuplicateLogin(user.Login)
		}
		r.log.Errorf("Database error during user update for ID %d: %v", user.ID, err)
		return domain.ErrDatabase(fmt.Errorf("failed to execute user update statement: %w", err))
	}

	count, err := res.RowsAffected()
	if err != nil {
		r.log.Errorf("Failed to get affected rows on user update for ID %d: %v", user.ID, err)
		return domain.ErrDatabase(fmt.Errorf("failed to get affected rows on user update: %w", err))
	}
	if count == 0 {
		r.log.Warnf("User update failed: User with ID %d not found.", user.ID)
		return domain.ErrNotFound("User", user.ID)
	}

	r.log.Infof("Successfully updated user with ID %d", user.ID)
	return nil
}

// Delete removes a user from the database.
func (r *sqlUserRepository) Delete(id int64) error {
	r.log.Debugf("Attempting to delete user with ID: %d", id)
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		r.log.Errorf("Database error during user delete for ID %d: %v", id, err)
		return domain.ErrDatabase(fmt.Errorf("failed to execute user delete statement: %w", err))
	}

	count, err := res.RowsAffected()
	if err != nil {
		r.log.Errorf("Failed to get affected rows on user delete for ID %d: %v", id, err)
		return domain.ErrDatabase(fmt.Errorf("failed to get affected rows on user delete: %w", err))
	}
	if count == 0 {
		r.log.Warnf("User delete failed: User with ID %d not found.", id)
		return domain.ErrNotFound("User", id) // User with that ID didn't exist
	}
	r.log.Infof("Successfully deleted user with ID %d", id)
	return nil
}
