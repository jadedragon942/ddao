package main

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/jadedragon942/ddao/orm"
)

type LDAPServer struct {
	port     int
	baseDN   string
	bindDN   string
	bindPW   string
	orm      *orm.ORM
	verbose  bool
	listener net.Listener
	quit     chan bool
}

func NewLDAPServer(port int, baseDN, bindDN, bindPW string, ormInstance *orm.ORM, verbose bool) *LDAPServer {
	return &LDAPServer{
		port:    port,
		baseDN:  baseDN,
		bindDN:  bindDN,
		bindPW:  bindPW,
		orm:     ormInstance,
		verbose: verbose,
		quit:    make(chan bool),
	}
}

func (s *LDAPServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	s.listener = listener

	for {
		select {
		case <-s.quit:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

func (s *LDAPServer) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *LDAPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	if s.verbose {
		log.Printf("New connection from %s", conn.RemoteAddr())
	}

	// This is a simplified LDAP server implementation
	// In a real implementation, you would need to properly parse LDAP messages
	// and implement the full LDAP protocol

	// For demonstration purposes, we'll implement basic LDAP operations
	// using a simple text-based protocol

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if s.verbose {
				log.Printf("Connection closed: %v", err)
			}
			return
		}

		message := string(buffer[:n])
		if s.verbose {
			log.Printf("Received: %s", message)
		}

		response := s.processMessage(message)

		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}
	}
}

func (s *LDAPServer) processMessage(message string) string {
	// Simple command processing for demonstration
	// In a real LDAP server, you would parse BER-encoded LDAP messages

	parts := strings.Fields(strings.TrimSpace(message))
	if len(parts) == 0 {
		return "ERROR: Empty message\n"
	}

	command := strings.ToUpper(parts[0])

	switch command {
	case "BIND":
		if len(parts) < 3 {
			return "ERROR: BIND requires DN and password\n"
		}
		return s.handleBind(parts[1], strings.Join(parts[2:], " "))

	case "SEARCH":
		if len(parts) < 2 {
			return "ERROR: SEARCH requires base DN\n"
		}
		filter := ""
		if len(parts) > 2 {
			filter = strings.Join(parts[2:], " ")
		}
		return s.handleSearch(parts[1], filter)

	case "ADD":
		if len(parts) < 2 {
			return "ERROR: ADD requires DN\n"
		}
		return s.handleAdd(parts[1], strings.Join(parts[2:], " "))

	case "DELETE":
		if len(parts) < 2 {
			return "ERROR: DELETE requires DN\n"
		}
		return s.handleDelete(parts[1])

	case "MODIFY":
		if len(parts) < 2 {
			return "ERROR: MODIFY requires DN\n"
		}
		return s.handleModify(parts[1], strings.Join(parts[2:], " "))

	case "HELP":
		return s.getHelp()

	default:
		return fmt.Sprintf("ERROR: Unknown command: %s\n", command)
	}
}

func (s *LDAPServer) handleBind(dn, password string) string {
	// Check if this is the admin bind
	if dn == s.bindDN && password == s.bindPW {
		return "OK: Admin bind successful\n"
	}

	// Check user authentication
	authenticated, err := s.authenticateUser(dn, password)
	if err != nil {
		return fmt.Sprintf("ERROR: Authentication failed: %v\n", err)
	}

	if authenticated {
		return "OK: Bind successful\n"
	}

	return "ERROR: Invalid credentials\n"
}

func (s *LDAPServer) handleSearch(baseDN, filter string) string {
	entries, err := s.searchEntries(baseDN, filter)
	if err != nil {
		return fmt.Sprintf("ERROR: Search failed: %v\n", err)
	}

	if len(entries) == 0 {
		return "OK: No entries found\n"
	}

	result := "OK: Search results:\n"
	for _, entry := range entries {
		result += fmt.Sprintf("DN: %s\n", entry.DN)
		result += fmt.Sprintf("ObjectClass: %s\n", entry.ObjectClass)
		result += fmt.Sprintf("Attributes: %s\n", entry.Attributes)
		result += "---\n"
	}

	return result
}

func (s *LDAPServer) handleAdd(dn, attributes string) string {
	err := s.addEntry(dn, attributes)
	if err != nil {
		return fmt.Sprintf("ERROR: Add failed: %v\n", err)
	}
	return "OK: Entry added\n"
}

func (s *LDAPServer) handleDelete(dn string) string {
	err := s.deleteEntry(dn)
	if err != nil {
		return fmt.Sprintf("ERROR: Delete failed: %v\n", err)
	}
	return "OK: Entry deleted\n"
}

func (s *LDAPServer) handleModify(dn, attributes string) string {
	err := s.modifyEntry(dn, attributes)
	if err != nil {
		return fmt.Sprintf("ERROR: Modify failed: %v\n", err)
	}
	return "OK: Entry modified\n"
}

func (s *LDAPServer) getHelp() string {
	return `LDAP Server Commands:
BIND <dn> <password>     - Authenticate with DN and password
SEARCH <base> [filter]   - Search for entries
ADD <dn> <attributes>    - Add new entry
DELETE <dn>              - Delete entry
MODIFY <dn> <attributes> - Modify entry
HELP                     - Show this help

Example:
BIND cn=admin,dc=example,dc=com admin
SEARCH dc=example,dc=com
ADD cn=john,ou=users,dc=example,dc=com objectClass=person givenName=John sn=Doe
`
}

// Helper function to hash passwords
func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}

// TLS support (optional)
func (s *LDAPServer) enableTLS(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.port), config)
	if err != nil {
		return err
	}

	s.listener = listener
	return nil
}