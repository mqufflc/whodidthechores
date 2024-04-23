package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mqufflc/whodidthechores/internal/utils"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", nil
	}
	return string(hashedPassword), nil
}

func verifyPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func checkPasswordComplexity(password string) error {
	const minEntropyBits = 65
	return passwordvalidator.Validate(password, minEntropyBits)
}

func (service *Service) createSession(user User) (*Session, error) {
	id := uuid.New()
	session, err := service.CreateSession(SessionParams{ID: id.String(), UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour * 4)})
	if err != nil {
		slog.Error("Error while creating session")
		return session, err
	}
	slog.Info(fmt.Sprintf("%v", session))
	return session, nil
}

func (service *Service) Login(creds Credentials) (*Session, error) {
	user, err := service.SearchUserByName(creds.Name)
	if user == nil || err != nil {
		return &Session{}, errors.New("username not found")
	}

	if err = verifyPassword(creds.Password, user.Hash); err != nil {
		return &Session{}, errors.New("bad password")
	}

	return service.createSession(*user)

}

func (service *Service) SignUp(creds Credentials) (*Session, error) {
	err := checkPasswordComplexity(creds.Password)

	if err != nil {
		return &Session{}, err
	}

	hash, err := HashPassword(creds.Password)

	if err != nil {
		return &Session{}, errors.New("unable to generate password hash")
	}

	slog.Info("Credentials received : %v, %v", creds.Name, creds.Password)

	user, err := service.CreateUser(UserParams{ID: utils.GenerateBase58ID(10), Name: creds.Name, Hash: hash})
	if user == nil || err != nil {
		return &Session{}, err
	}

	return service.createSession(*user)
}
