package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jadedragon942/ddao/orm"
)

type RestdService struct {
	orm *orm.ORM
}

func NewRestdService(orm *orm.ORM) *RestdService {
	return &RestdService{orm: orm}
}

// User handlers

func (s *RestdService) CreateUser(ctx context.Context, input *CreateUserInput) (*UserResponse, error) {
	user := User{
		ID:        generateID("user"),
		Email:     input.Body.Email,
		Name:      input.Body.Name,
		Profile:   input.Body.Profile,
		CreatedAt: time.Now(),
	}

	obj := userToObject(user)
	_, created, err := s.orm.Insert(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	if !created {
		return nil, fmt.Errorf("user with email %s already exists", user.Email)
	}

	return &UserResponse{Body: user}, nil
}

func (s *RestdService) GetUser(ctx context.Context, input *struct {
	UserID string `path:"userId" example:"user123" doc:"User ID"`
}) (*UserResponse, error) {
	obj, err := s.orm.FindByID(ctx, "users", input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("user not found")
	}

	user := objectToUser(obj)
	return &UserResponse{Body: user}, nil
}

func (s *RestdService) GetUserByEmail(ctx context.Context, input *struct {
	Email string `query:"email" example:"user@example.com" doc:"User email"`
}) (*UserResponse, error) {
	obj, err := s.orm.FindByKey(ctx, "users", "email", input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("user not found")
	}

	user := objectToUser(obj)
	return &UserResponse{Body: user}, nil
}

func (s *RestdService) UpdateUser(ctx context.Context, input *UpdateUserInput) (*UserResponse, error) {
	obj, err := s.orm.FindByID(ctx, "users", input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("user not found")
	}

	now := time.Now()
	if input.Body.Email != nil {
		obj.SetField("email", *input.Body.Email)
	}
	if input.Body.Name != nil {
		obj.SetField("name", *input.Body.Name)
	}
	if input.Body.Profile != nil {
		profileJSON, _ := json.Marshal(input.Body.Profile)
		obj.SetField("profile", string(profileJSON))
	}
	obj.SetField("updated_at", now.Format(time.RFC3339))

	_, err = s.orm.Storage.Update(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user := objectToUser(obj)
	return &UserResponse{Body: user}, nil
}

func (s *RestdService) DeleteUser(ctx context.Context, input *struct {
	UserID string `path:"userId" example:"user123" doc:"User ID"`
}) (*DeleteResponse, error) {
	deleted, err := s.orm.DeleteByID(ctx, "users", input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}
	if !deleted {
		return nil, fmt.Errorf("user not found")
	}

	return &DeleteResponse{
		Body: struct {
			Success bool   `json:"success" example:"true" doc:"Whether the deletion was successful"`
			Message string `json:"message" example:"User deleted successfully" doc:"Success message"`
		}{
			Success: true,
			Message: "User deleted successfully",
		},
	}, nil
}

// Post handlers

func (s *RestdService) CreatePost(ctx context.Context, input *CreatePostInput) (*PostResponse, error) {
	// Verify user exists
	userObj, err := s.orm.FindByID(ctx, "users", input.Body.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if userObj == nil {
		return nil, fmt.Errorf("user not found")
	}

	published := false
	if input.Body.Published != nil {
		published = *input.Body.Published
	}

	post := Post{
		ID:        generateID("post"),
		UserID:    input.Body.UserID,
		Title:     input.Body.Title,
		Content:   input.Body.Content,
		Metadata:  input.Body.Metadata,
		Published: published,
		CreatedAt: time.Now(),
	}

	obj := postToObject(post)
	_, created, err := s.orm.Insert(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	if !created {
		return nil, fmt.Errorf("post with ID %s already exists", post.ID)
	}

	return &PostResponse{Body: post}, nil
}

func (s *RestdService) GetPost(ctx context.Context, input *struct {
	PostID string `path:"postId" example:"post123" doc:"Post ID"`
}) (*PostResponse, error) {
	obj, err := s.orm.FindByID(ctx, "posts", input.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("post not found")
	}

	post := objectToPost(obj)
	return &PostResponse{Body: post}, nil
}

func (s *RestdService) UpdatePost(ctx context.Context, input *UpdatePostInput) (*PostResponse, error) {
	obj, err := s.orm.FindByID(ctx, "posts", input.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("post not found")
	}

	now := time.Now()
	if input.Body.Title != nil {
		obj.SetField("title", *input.Body.Title)
	}
	if input.Body.Content != nil {
		obj.SetField("content", *input.Body.Content)
	}
	if input.Body.Metadata != nil {
		metadataJSON, _ := json.Marshal(input.Body.Metadata)
		obj.SetField("metadata", string(metadataJSON))
	}
	if input.Body.Published != nil {
		obj.SetField("published", *input.Body.Published)
	}
	obj.SetField("updated_at", now.Format(time.RFC3339))

	_, err = s.orm.Storage.Update(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	post := objectToPost(obj)
	return &PostResponse{Body: post}, nil
}

func (s *RestdService) DeletePost(ctx context.Context, input *struct {
	PostID string `path:"postId" example:"post123" doc:"Post ID"`
}) (*DeleteResponse, error) {
	deleted, err := s.orm.DeleteByID(ctx, "posts", input.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete post: %w", err)
	}
	if !deleted {
		return nil, fmt.Errorf("post not found")
	}

	return &DeleteResponse{
		Body: struct {
			Success bool   `json:"success" example:"true" doc:"Whether the deletion was successful"`
			Message string `json:"message" example:"User deleted successfully" doc:"Success message"`
		}{
			Success: true,
			Message: "Post deleted successfully",
		},
	}, nil
}

// Helper function to generate IDs
func generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}