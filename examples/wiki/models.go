package main

import (
	"time"

	"github.com/jadedragon942/ddao/object"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type WikiPage struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func userToObject(user *User) *object.Object {
	obj := object.New()
	obj.TableName = "users"
	obj.ID = user.ID
	obj.Fields = map[string]any{
		"username":   user.Username,
		"email":      user.Email,
		"password":   user.Password,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	}
	return obj
}

func objectToUser(obj *object.Object) *User {
	user := &User{
		ID: obj.ID,
	}

	if username, exists := obj.GetString("username"); exists {
		user.Username = username
	}
	if email, exists := obj.GetString("email"); exists {
		user.Email = email
	}
	if password, exists := obj.GetString("password"); exists {
		user.Password = password
	}
	if createdAtStr, exists := obj.GetString("created_at"); exists {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			user.CreatedAt = t
		}
	}

	return user
}

func wikiPageToObject(page *WikiPage) *object.Object {
	obj := object.New()
	obj.TableName = "wiki_pages"
	obj.ID = page.ID
	obj.Fields = map[string]any{
		"title":      page.Title,
		"content":    page.Content,
		"author_id":  page.AuthorID,
		"created_at": page.CreatedAt.Format(time.RFC3339),
		"updated_at": page.UpdatedAt.Format(time.RFC3339),
	}
	return obj
}

func objectToWikiPage(obj *object.Object) *WikiPage {
	page := &WikiPage{
		ID: obj.ID,
	}

	if title, exists := obj.GetString("title"); exists {
		page.Title = title
	}
	if content, exists := obj.GetString("content"); exists {
		page.Content = content
	}
	if authorID, exists := obj.GetString("author_id"); exists {
		page.AuthorID = authorID
	}
	if createdAtStr, exists := obj.GetString("created_at"); exists {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			page.CreatedAt = t
		}
	}
	if updatedAtStr, exists := obj.GetString("updated_at"); exists {
		if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			page.UpdatedAt = t
		}
	}

	return page
}

func sessionToObject(session *Session) *object.Object {
	obj := object.New()
	obj.TableName = "sessions"
	obj.ID = session.ID
	obj.Fields = map[string]any{
		"user_id":    session.UserID,
		"created_at": session.CreatedAt.Format(time.RFC3339),
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	}
	return obj
}

func objectToSession(obj *object.Object) *Session {
	session := &Session{
		ID: obj.ID,
	}

	if userID, exists := obj.GetString("user_id"); exists {
		session.UserID = userID
	}
	if createdAtStr, exists := obj.GetString("created_at"); exists {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			session.CreatedAt = t
		}
	}
	if expiresAtStr, exists := obj.GetString("expires_at"); exists {
		if t, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
			session.ExpiresAt = t
		}
	}

	return session
}