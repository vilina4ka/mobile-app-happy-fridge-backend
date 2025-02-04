package store

import "happy-fridge-rest-api/cmd/internal/app/model"

// UserRepository ...
type UserRepository struct {
	store *Store
}

// Create ...
func (r *UserRepository) Create(u *model.User) (*model.User, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}
	if err := r.store.db.QueryRow("INSERT INTO users (email, encrypted_password) VALUES ($1, $2) RETURNING id, fridge_number", u.Email, u.EncryptedPassword).Scan(&u.ID, &u.FridgeNumber); err != nil {
		return nil, err
	}
	return u, nil
}

// FindByEmail ...
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, email, encrypted_password, fridge_number FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.EncryptedPassword, &u.FridgeNumber); err != nil {
		return nil, err
	}
	return u, nil
}

// FindByFridgeNumber ...
func (r *UserRepository) FindByFridgeNumber(fridgeNumber string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, email, encrypted_password, fridge_number FROM users WHERE fridge_number = $1", fridgeNumber).
		Scan(&u.ID, &u.Email, &u.EncryptedPassword, &u.FridgeNumber); err != nil {
		return nil, err
	}
	return u, nil
}
