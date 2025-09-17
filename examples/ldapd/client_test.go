package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// TestClient provides a simple test client for the LDAP server
type TestClient struct {
	conn net.Conn
}

func NewTestClient(address string) (*TestClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TestClient{conn: conn}, nil
}

func (c *TestClient) Close() error {
	return c.conn.Close()
}

func (c *TestClient) SendCommand(command string) (string, error) {
	// Send command
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	// Read response
	buffer := make([]byte, 4096)
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := c.conn.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer[:n]), nil
}

// Integration test - requires server to be running
func TestLDAPServerIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Connect to server (assumes server is running on localhost:1389)
	client, err := NewTestClient("localhost:1389")
	if err != nil {
		t.Skipf("Could not connect to LDAP server: %v", err)
	}
	defer client.Close()

	// Test admin bind
	t.Run("AdminBind", func(t *testing.T) {
		response, err := client.SendCommand("BIND cn=admin,dc=example,dc=com admin")
		if err != nil {
			t.Fatalf("Failed to send bind command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test search
	t.Run("Search", func(t *testing.T) {
		response, err := client.SendCommand("SEARCH dc=example,dc=com")
		if err != nil {
			t.Fatalf("Failed to send search command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test add user
	t.Run("AddUser", func(t *testing.T) {
		response, err := client.SendCommand("ADD cn=testuser,ou=users,dc=example,dc=com objectClass=person givenName=Test sn=User userPassword=testpass")
		if err != nil {
			t.Fatalf("Failed to send add command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test user authentication
	t.Run("UserBind", func(t *testing.T) {
		response, err := client.SendCommand("BIND cn=testuser,ou=users,dc=example,dc=com testpass")
		if err != nil {
			t.Fatalf("Failed to send bind command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test modify user
	t.Run("ModifyUser", func(t *testing.T) {
		response, err := client.SendCommand("MODIFY cn=testuser,ou=users,dc=example,dc=com description=Modified")
		if err != nil {
			t.Fatalf("Failed to send modify command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test delete user
	t.Run("DeleteUser", func(t *testing.T) {
		response, err := client.SendCommand("DELETE cn=testuser,ou=users,dc=example,dc=com")
		if err != nil {
			t.Fatalf("Failed to send delete command: %v", err)
		}
		if !containsOK(response) {
			t.Errorf("Expected OK response, got: %s", response)
		}
	})

	// Test help command
	t.Run("Help", func(t *testing.T) {
		response, err := client.SendCommand("HELP")
		if err != nil {
			t.Fatalf("Failed to send help command: %v", err)
		}
		if !contains(response, "LDAP Server Commands") {
			t.Errorf("Expected help response, got: %s", response)
		}
	})
}

func containsOK(response string) bool {
	return contains(response, "OK:")
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		   fmt.Sprintf("%s", str[:len(substr)]) == substr ||
		   len(str) > len(substr) &&
		   fmt.Sprintf("%s", str[len(str)-len(substr):]) == substr ||
		   len(str) > len(substr) &&
		   fmt.Sprintf("%v", str) != str // fallback check
}

// Benchmark test for server performance
func BenchmarkLDAPOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client, err := NewTestClient("localhost:1389")
	if err != nil {
		b.Skipf("Could not connect to LDAP server: %v", err)
	}
	defer client.Close()

	// Bind once
	_, err = client.SendCommand("BIND cn=admin,dc=example,dc=com admin")
	if err != nil {
		b.Fatalf("Failed to bind: %v", err)
	}

	b.Run("Search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.SendCommand("SEARCH dc=example,dc=com")
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})

	b.Run("AddDelete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dn := fmt.Sprintf("cn=bench%d,ou=users,dc=example,dc=com", i)

			// Add
			_, err := client.SendCommand(fmt.Sprintf("ADD %s objectClass=person", dn))
			if err != nil {
				b.Fatalf("Add failed: %v", err)
			}

			// Delete
			_, err = client.SendCommand(fmt.Sprintf("DELETE %s", dn))
			if err != nil {
				b.Fatalf("Delete failed: %v", err)
			}
		}
	})
}