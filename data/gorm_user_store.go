package data

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"gorm.io/gorm"
)

type gormUserStore struct {
	DB *gorm.DB
}

func (s gormUserStore) GetUserIdentifierById(userId uint64) (*UserIdentifier, error) {
	var user UserIdentifier
	err := s.DB.Model(&User{}).First(&user, userId).Error
	if err != nil {
		return nil, UserNotFoundError{ID: userId}
	}
	return &user, nil
}

func (s gormUserStore) GetUserPersonalById(userId uint64) (*UserPersonal, error) {
	var user UserPersonal
	err := s.DB.Model(&User{}).First(&user, userId).Error
	if err != nil {
		return nil, UserNotFoundError{ID: userId}
	}
	return &user, nil
}

func (s gormUserStore) GetUserByEmail(email string) (*User, error) {
	var user User
	err := s.DB.Where(User{
		Email: email,
	}).First(&user).Error
	if err != nil {
		return nil, UserNotFoundError{Email: email}
	}
	return &user, nil
}

func (s gormUserStore) CreateUser(user User) error {
	err := s.DB.Create(&user).Error

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return UserUniquenessError{
				User:       user,
				ErrorField: pgErr.ColumnName,
			}
		}
		return UserCreationError{
			User: user,
			Err:  pgErr,
		}
	}

	return nil
}

func initGormUserStore(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}

	Users = &gormUserStore{
		DB: db,
	}
	return nil
}
