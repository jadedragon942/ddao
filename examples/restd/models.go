package main

import (
	"encoding/json"
	"time"

	"github.com/jadedragon942/ddao/object"
)

// User represents a user in the system
type User struct {
	ID        string                 `json:"id" example:"user123" doc:"User ID"`
	Email     string                 `json:"email" example:"user@example.com" doc:"User email address"`
	Name      string                 `json:"name" example:"John Doe" doc:"User full name"`
	Profile   map[string]interface{} `json:"profile,omitempty" doc:"User profile data"`
	CreatedAt time.Time              `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt *time.Time             `json:"updated_at,omitempty" doc:"Last update timestamp"`
}

// Post represents a post in the system
type Post struct {
	ID        string                 `json:"id" example:"post123" doc:"Post ID"`
	UserID    string                 `json:"user_id" example:"user123" doc:"User ID who created the post"`
	Title     string                 `json:"title" example:"My First Post" doc:"Post title"`
	Content   string                 `json:"content" example:"This is my first post content." doc:"Post content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" doc:"Post metadata"`
	Published bool                   `json:"published" example:"false" doc:"Whether the post is published"`
	CreatedAt time.Time              `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt *time.Time             `json:"updated_at,omitempty" doc:"Last update timestamp"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Body struct {
		Email   string                 `json:"email" required:"true" example:"user@example.com" doc:"User email address"`
		Name    string                 `json:"name" required:"true" example:"John Doe" doc:"User full name"`
		Profile map[string]interface{} `json:"profile,omitempty" doc:"User profile data"`
	}
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	UserID string `path:"userId" example:"user123" doc:"User ID to update"`
	Body   struct {
		Email   *string                `json:"email,omitempty" example:"user@example.com" doc:"User email address"`
		Name    *string                `json:"name,omitempty" example:"John Doe" doc:"User full name"`
		Profile map[string]interface{} `json:"profile,omitempty" doc:"User profile data"`
	}
}

// CreatePostInput represents input for creating a post
type CreatePostInput struct {
	Body struct {
		UserID    string                 `json:"user_id" required:"true" example:"user123" doc:"User ID who creates the post"`
		Title     string                 `json:"title" required:"true" example:"My First Post" doc:"Post title"`
		Content   string                 `json:"content" required:"true" example:"This is my first post content." doc:"Post content"`
		Metadata  map[string]interface{} `json:"metadata,omitempty" doc:"Post metadata"`
		Published *bool                  `json:"published,omitempty" example:"false" doc:"Whether the post is published"`
	}
}

// UpdatePostInput represents input for updating a post
type UpdatePostInput struct {
	PostID string `path:"postId" example:"post123" doc:"Post ID to update"`
	Body   struct {
		Title     *string                `json:"title,omitempty" example:"My Updated Post" doc:"Post title"`
		Content   *string                `json:"content,omitempty" example:"This is updated content." doc:"Post content"`
		Metadata  map[string]interface{} `json:"metadata,omitempty" doc:"Post metadata"`
		Published *bool                  `json:"published,omitempty" example:"true" doc:"Whether the post is published"`
	}
}

// UserResponse represents response containing a user
type UserResponse struct {
	Body User `json:"user"`
}

// PostResponse represents response containing a post
type PostResponse struct {
	Body Post `json:"post"`
}

// UsersResponse represents response containing multiple users
type UsersResponse struct {
	Body struct {
		Users []User `json:"users" doc:"List of users"`
		Count int    `json:"count" doc:"Total number of users"`
	}
}

// PostsResponse represents response containing multiple posts
type PostsResponse struct {
	Body struct {
		Posts []Post `json:"posts" doc:"List of posts"`
		Count int    `json:"count" doc:"Total number of posts"`
	}
}

// DeleteResponse represents a successful deletion response
type DeleteResponse struct {
	Body struct {
		Success bool   `json:"success" example:"true" doc:"Whether the deletion was successful"`
		Message string `json:"message" example:"User deleted successfully" doc:"Success message"`
	}
}

// Helper functions to convert between models and objects

func userToObject(user User) *object.Object {
	obj := object.New()
	obj.TableName = "users"
	obj.ID = user.ID
	obj.Fields = map[string]any{
		"email":      user.Email,
		"name":       user.Name,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	}

	if user.Profile != nil {
		profileJSON, _ := json.Marshal(user.Profile)
		obj.Fields["profile"] = string(profileJSON)
	}

	if user.UpdatedAt != nil {
		obj.Fields["updated_at"] = user.UpdatedAt.Format(time.RFC3339)
	}

	return obj
}

func objectToUser(obj *object.Object) User {
	user := User{
		ID: obj.ID,
	}

	if email, exists := obj.GetString("email"); exists {
		user.Email = email
	}
	if name, exists := obj.GetString("name"); exists {
		user.Name = name
	}
	if createdAtStr, exists := obj.GetString("created_at"); exists {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			user.CreatedAt = t
		}
	}
	if updatedAtStr, exists := obj.GetString("updated_at"); exists && updatedAtStr != "" {
		if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			user.UpdatedAt = &t
		}
	}
	if profileStr, exists := obj.GetString("profile"); exists && profileStr != "" {
		var profile map[string]interface{}
		if err := json.Unmarshal([]byte(profileStr), &profile); err == nil {
			user.Profile = profile
		}
	}

	return user
}

func postToObject(post Post) *object.Object {
	obj := object.New()
	obj.TableName = "posts"
	obj.ID = post.ID
	obj.Fields = map[string]any{
		"user_id":    post.UserID,
		"title":      post.Title,
		"content":    post.Content,
		"published":  post.Published,
		"created_at": post.CreatedAt.Format(time.RFC3339),
	}

	if post.Metadata != nil {
		metadataJSON, _ := json.Marshal(post.Metadata)
		obj.Fields["metadata"] = string(metadataJSON)
	}

	if post.UpdatedAt != nil {
		obj.Fields["updated_at"] = post.UpdatedAt.Format(time.RFC3339)
	}

	return obj
}

func objectToPost(obj *object.Object) Post {
	post := Post{
		ID: obj.ID,
	}

	if userID, exists := obj.GetString("user_id"); exists {
		post.UserID = userID
	}
	if title, exists := obj.GetString("title"); exists {
		post.Title = title
	}
	if content, exists := obj.GetString("content"); exists {
		post.Content = content
	}
	if published, exists := obj.GetBool("published"); exists {
		post.Published = published
	}
	if createdAtStr, exists := obj.GetString("created_at"); exists {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			post.CreatedAt = t
		}
	}
	if updatedAtStr, exists := obj.GetString("updated_at"); exists && updatedAtStr != "" {
		if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			post.UpdatedAt = &t
		}
	}
	if metadataStr, exists := obj.GetString("metadata"); exists && metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			post.Metadata = metadata
		}
	}

	return post
}