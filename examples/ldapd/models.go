package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jadedragon942/ddao/object"
	"github.com/jadedragon942/ddao/orm"
)

// LDAPEntry represents an LDAP directory entry
type LDAPEntry struct {
	DN          string `json:"dn"`
	ParentDN    string `json:"parent_dn,omitempty"`
	ObjectClass string `json:"object_class"`
	Attributes  string `json:"attributes"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// LDAPUser represents a user for authentication
type LDAPUser struct {
	DN           string    `json:"dn"`
	PasswordHash string    `json:"password_hash"`
	Salt         string    `json:"salt"`
	LastLogin    time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// LDAPGroup represents group membership
type LDAPGroup struct {
	ID        string    `json:"id"`
	GroupDN   string    `json:"group_dn"`
	MemberDN  string    `json:"member_dn"`
	CreatedAt time.Time `json:"created_at"`
}

// LDAP operations using DDAO

func (s *LDAPServer) searchEntries(baseDN, filter string) ([]LDAPEntry, error) {
	ctx := context.Background()
	var entries []LDAPEntry

	// For this example, we'll do a simple search by object class or DN
	// In a real implementation, you would parse LDAP filters properly

	if filter == "" || strings.Contains(filter, "objectClass") {
		// Search all entries under base DN
		// This is a simplified implementation - in practice you'd need proper LDAP filter parsing

		// For demonstration, let's just return entries that start with the base DN
		// In a real implementation, you'd implement proper hierarchical searching

		// Since we can't do complex queries easily with DDAO's simple interface,
		// we'll just search for specific patterns

		entry, err := s.orm.FindByID(ctx, "entries", baseDN)
		if err == nil && entry != nil {
			ldapEntry := LDAPEntry{
				DN:          entry.ID,
				ObjectClass: entry.Fields["object_class"].(string),
				Attributes:  entry.Fields["attributes"].(string),
				CreatedAt:   entry.Fields["created_at"].(string),
			}
			if parentDN, exists := entry.Fields["parent_dn"]; exists && parentDN != nil {
				ldapEntry.ParentDN = parentDN.(string)
			}
			if updatedAt, exists := entry.Fields["updated_at"]; exists && updatedAt != nil {
				ldapEntry.UpdatedAt = updatedAt.(string)
			}
			entries = append(entries, ldapEntry)
		}
	}

	return entries, nil
}

func (s *LDAPServer) addEntry(dn, attributesStr string) error {
	ctx := context.Background()

	// Parse attributes (simplified)
	attributes := parseAttributes(attributesStr)

	// Get object class
	objectClass := "person" // default
	if oc, exists := attributes["objectClass"]; exists {
		objectClass = oc
	}

	// Get parent DN
	parentDN := getParentDN(dn)

	// Create entry object
	entry := object.New()
	entry.TableName = "entries"
	entry.ID = dn
	entry.Fields = map[string]any{
		"parent_dn":    parentDN,
		"object_class": objectClass,
		"attributes":   attributesStr,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	_, created, err := s.orm.Insert(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	if !created {
		return fmt.Errorf("entry with DN %s already exists", dn)
	}

	// If this is a person with a password, create user record
	if objectClass == "person" || objectClass == "inetOrgPerson" {
		if password, exists := attributes["userPassword"]; exists {
			err := s.createUser(dn, password)
			if err != nil {
				return fmt.Errorf("failed to create user record: %w", err)
			}
		}
	}

	return nil
}

func (s *LDAPServer) deleteEntry(dn string) error {
	ctx := context.Background()

	// Delete from entries table
	deleted, err := s.orm.DeleteByID(ctx, "entries", dn)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	if !deleted {
		return fmt.Errorf("entry with DN %s not found", dn)
	}

	// Also delete user record if exists
	s.orm.DeleteByID(ctx, "users", dn)

	// Remove from groups
	// This is simplified - in practice you'd need better group management

	return nil
}

func (s *LDAPServer) modifyEntry(dn, attributesStr string) error {
	ctx := context.Background()

	// Find existing entry
	entry, err := s.orm.FindByID(ctx, "entries", dn)
	if err != nil {
		return fmt.Errorf("failed to find entry: %w", err)
	}
	if entry == nil {
		return fmt.Errorf("entry with DN %s not found", dn)
	}

	// Update attributes
	entry.Fields["attributes"] = attributesStr
	entry.Fields["updated_at"] = time.Now().Format(time.RFC3339)

	updated, err := s.orm.Storage.Update(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	if !updated {
		return fmt.Errorf("failed to update entry %s", dn)
	}

	return nil
}

func (s *LDAPServer) createUser(dn, password string) error {
	ctx := context.Background()

	// Generate salt
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)

	// Hash password
	passwordHash := hashPassword(password, salt)

	// Create user object
	user := object.New()
	user.TableName = "users"
	user.ID = dn
	user.Fields = map[string]any{
		"password_hash": passwordHash,
		"salt":          salt,
		"created_at":    time.Now().Format(time.RFC3339),
	}

	_, created, err := s.orm.Insert(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	if !created {
		return fmt.Errorf("user with DN %s already exists", dn)
	}

	return nil
}

func (s *LDAPServer) authenticateUser(dn, password string) (bool, error) {
	ctx := context.Background()

	// Find user
	user, err := s.orm.FindByID(ctx, "users", dn)
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	// Get stored hash and salt
	storedHash := user.Fields["password_hash"].(string)
	salt := user.Fields["salt"].(string)

	// Hash provided password
	providedHash := hashPassword(password, salt)

	// Compare hashes
	if storedHash == providedHash {
		// Update last login
		user.Fields["last_login"] = time.Now().Format(time.RFC3339)
		s.orm.Storage.Update(ctx, user)
		return true, nil
	}

	return false, nil
}

// Helper functions

func parseAttributes(attributesStr string) map[string]string {
	attributes := make(map[string]string)

	// Simple attribute parsing: "key=value key2=value2"
	parts := strings.Fields(attributesStr)
	for _, part := range parts {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			attributes[kv[0]] = kv[1]
		}
	}

	return attributes
}

func getParentDN(dn string) string {
	// Extract parent DN by removing the first component
	parts := strings.SplitN(dn, ",", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// Setup initial data for the LDAP server
func setupInitialData(ctx context.Context, ormInstance *orm.ORM, baseDN, bindDN, bindPW string) error {
	// Check if base DN already exists
	entry, err := ormInstance.FindByID(ctx, "entries", baseDN)
	if err == nil && entry != nil {
		// Base DN already exists, skip initialization
		return nil
	}

	// Create base DN entry
	baseEntry := object.New()
	baseEntry.TableName = "entries"
	baseEntry.ID = baseDN
	baseEntry.Fields = map[string]any{
		"object_class": "organization",
		"attributes":   fmt.Sprintf(`{"o": "Example Organization", "dc": "%s"}`, strings.Split(baseDN, "=")[1]),
		"created_at":   time.Now().Format(time.RFC3339),
	}

	_, _, err = ormInstance.Insert(ctx, baseEntry)
	if err != nil {
		return fmt.Errorf("failed to create base DN: %w", err)
	}

	// Create admin user
	server := &LDAPServer{orm: ormInstance}
	err = server.createUser(bindDN, bindPW)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create admin entry
	adminEntry := object.New()
	adminEntry.TableName = "entries"
	adminEntry.ID = bindDN
	adminEntry.Fields = map[string]any{
		"parent_dn":    baseDN,
		"object_class": "person",
		"attributes":   `{"cn": "admin", "userPassword": "admin"}`,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	_, _, err = ormInstance.Insert(ctx, adminEntry)
	if err != nil {
		return fmt.Errorf("failed to create admin entry: %w", err)
	}

	// Create organizational units
	ouUsers := fmt.Sprintf("ou=users,%s", baseDN)
	ouUsersEntry := object.New()
	ouUsersEntry.TableName = "entries"
	ouUsersEntry.ID = ouUsers
	ouUsersEntry.Fields = map[string]any{
		"parent_dn":    baseDN,
		"object_class": "organizationalUnit",
		"attributes":   `{"ou": "users"}`,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	_, _, err = ormInstance.Insert(ctx, ouUsersEntry)
	if err != nil {
		return fmt.Errorf("failed to create users OU: %w", err)
	}

	ouGroups := fmt.Sprintf("ou=groups,%s", baseDN)
	ouGroupsEntry := object.New()
	ouGroupsEntry.TableName = "entries"
	ouGroupsEntry.ID = ouGroups
	ouGroupsEntry.Fields = map[string]any{
		"parent_dn":    baseDN,
		"object_class": "organizationalUnit",
		"attributes":   `{"ou": "groups"}`,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	_, _, err = ormInstance.Insert(ctx, ouGroupsEntry)
	if err != nil {
		return fmt.Errorf("failed to create groups OU: %w", err)
	}

	return nil
}
