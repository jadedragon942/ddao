package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/jadedragon942/ddao/orm"
)

type AuthService struct {
	orm *orm.ORM
}

func NewAuthService(orm *orm.ORM) *AuthService {
	return &AuthService{orm: orm}
}

func (a *AuthService) Register(username, email, password string) (*User, error) {
	userID, err := generateID()
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        userID,
		Username:  username,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	obj := userToObject(user)
	_, _, err = a.orm.Insert(context.Background(), obj)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (a *AuthService) Login(username, password string) (*User, string, error) {
	obj, err := a.orm.FindByKey(context.Background(), "users", "username", username)
	if err != nil {
		return nil, "", err
	}
	if obj == nil {
		return nil, "", errors.New("invalid username or password")
	}

	user := objectToUser(obj)
	if err := user.CheckPassword(password); err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	sessionID, err := a.CreateSession(user.ID)
	if err != nil {
		return nil, "", err
	}

	user.Password = ""
	return user, sessionID, nil
}

func (a *AuthService) CreateSession(userID string) (string, error) {
	sessionID, err := generateID()
	if err != nil {
		return "", err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	obj := sessionToObject(session)
	_, _, err = a.orm.Insert(context.Background(), obj)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (a *AuthService) ValidateSession(sessionID string) (*User, error) {
	obj, err := a.orm.FindByID(context.Background(), "sessions", sessionID)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("invalid session")
	}

	session := objectToSession(obj)
	if time.Now().After(session.ExpiresAt) {
		a.orm.DeleteByID(context.Background(), "sessions", sessionID)
		return nil, errors.New("session expired")
	}

	userObj, err := a.orm.FindByID(context.Background(), "users", session.UserID)
	if err != nil {
		return nil, err
	}
	if userObj == nil {
		return nil, errors.New("user not found")
	}

	user := objectToUser(userObj)
	user.Password = ""
	return user, nil
}

func (a *AuthService) Logout(sessionID string) error {
	_, err := a.orm.DeleteByID(context.Background(), "sessions", sessionID)
	return err
}

func (a *AuthService) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := a.ValidateSession(cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next(w, r.WithContext(ctx))
	}
}

func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}