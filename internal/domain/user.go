package domain

import (
	"encoding/base64"
	"fmt"
	"github.com/ougirez/diplom/internal/pkg/constants"
	"time"

	"github.com/google/uuid"
)

type UserName struct {
	FirstName string `json:"firstName" db:"first_name" validate:"required"`
	LastName  string `json:"lastName" db:"last_name" validate:"required"`
}

func (un *UserName) Full() string {
	return fmt.Sprintf("%s %s", un.FirstName, un.LastName)
}

type UserPassword struct {
	Hash string `json:"-" db:"password_hash"`
	Salt string `json:"-" db:"password_salt"`
}

type User struct {
	ID                    int64      `json:"id" db:"id"`
	Email                 NullString `json:"email" db:"email"`
	Image                 NullString `json:"image" db:"image"`
	SubscriptionExpiredAt *time.Time `json:"subscription_expired_at" db:"subscription_expired_at"`
	MainAccountNumber     uuid.UUID  `json:"mainAccountNumber" db:"main_account_number"`
	UserName
	UserPassword
}

type UserStatus string

const (
	UserStatusPendindApprove UserStatus = "pending_approve"
	UserStatusApproved       UserStatus = "approved"
	UserStatusBanned         UserStatus = "banned"
)

// Init generates salt and hash with given password and fills corresponding fields.
func (up *UserPassword) Init(password string) error {
	salt, err := GetSalt()
	if err != nil {
		return fmt.Errorf("error generating salt: %s", err)
	}

	hash, err := GetHash512(password, salt)
	if err != nil {
		return fmt.Errorf("error generating hash: %s", err)
	}

	up.Salt = base64.URLEncoding.EncodeToString(salt)
	up.Hash = base64.URLEncoding.EncodeToString(hash)

	return nil
}

// Validate checks if the given password is the one that is stored.
func (up *UserPassword) Validate(password string) error {
	salt, err := base64.URLEncoding.DecodeString(up.Salt)
	if err != nil {
		return fmt.Errorf("error decoding user's salt: %s", err)
	}

	hash, err := GetHash512(password, salt)
	if err != nil {
		return fmt.Errorf("error generating hash: %s", err)
	}

	if base64.URLEncoding.EncodeToString(hash) != up.Hash {
		return constants.ErrPasswordMismatch
	}

	return nil
}

type SignupUserRequest struct {
	UserName
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignupUserResponse struct {
	*User
	AuthToken string `json:"auth_token"`
}

type LoginUserRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponse struct {
	*User
	AuthToken string `json:"auth_token"`
}
