#!/bin/bash

# Simple test script for LDAP server

HOST="localhost"
PORT="1389"

echo "Testing LDAP server at $HOST:$PORT"
echo

# Function to send command and get response
send_command() {
    local cmd="$1"
    echo "Sending: $cmd"
    echo -n "$cmd" | timeout 2 telnet $HOST $PORT 2>/dev/null | tail -n +3
    echo
}

# Test commands
echo "=== Testing HELP command ==="
send_command "HELP"

echo "=== Testing BIND command ==="
send_command "BIND cn=admin,dc=example,dc=com admin"

echo "=== Testing SEARCH command ==="
send_command "SEARCH dc=example,dc=com"

echo "=== Testing ADD command ==="
send_command "ADD cn=testuser,ou=users,dc=example,dc=com objectClass=person givenName=Test sn=User userPassword=testpass"

echo "=== Testing USER BIND command ==="
send_command "BIND cn=testuser,ou=users,dc=example,dc=com testpass"

echo "=== Testing SEARCH for new user ==="
send_command "SEARCH cn=testuser,ou=users,dc=example,dc=com"

echo "=== Testing DELETE command ==="
send_command "DELETE cn=testuser,ou=users,dc=example,dc=com"

echo "Tests completed."