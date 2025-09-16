#!/bin/bash

# Demo script for RESTD - RESTful microservice with DDAO
# This script demonstrates the key features of the REST API

set -e

# Configuration
BASE_URL="http://localhost:8080"
API_PORT=8080

echo "🚀 RESTD Demo - RESTful microservice with DDAO"
echo "============================================="
echo

# Check if the server is running, start if not
if ! curl -s ${BASE_URL}/health > /dev/null 2>&1; then
    echo "📡 Starting RESTD server on port ${API_PORT}..."
    ./restd --storage-engine sqlite --connection-url ":memory:" --port ${API_PORT} &
    SERVER_PID=$!
    echo "   Server PID: ${SERVER_PID}"

    # Wait for server to start
    echo "   Waiting for server to start..."
    for i in {1..10}; do
        if curl -s ${BASE_URL}/health > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    if ! curl -s ${BASE_URL}/health > /dev/null 2>&1; then
        echo "❌ Failed to start server"
        exit 1
    fi

    echo "✅ Server started successfully!"
    echo

    # Setup cleanup trap
    trap 'echo "🛑 Stopping server..."; kill $SERVER_PID 2>/dev/null || true; exit' INT TERM EXIT
else
    echo "✅ Server is already running"
    echo
fi

# Check health
echo "🏥 Health Check"
echo "---------------"
HEALTH=$(curl -s ${BASE_URL}/health)
echo "Status: $(echo $HEALTH | jq -r '.status')"
echo "Storage Engine: $(echo $HEALTH | jq -r '.storage')"
echo

# Create a user
echo "👤 Creating a User"
echo "------------------"
USER_RESPONSE=$(curl -s -X POST ${BASE_URL}/users \
    -H "Content-Type: application/json" \
    -d '{
        "email": "demo@example.com",
        "name": "Demo User",
        "profile": {
            "location": "San Francisco",
            "bio": "Demo user for RESTD",
            "interests": ["technology", "api", "go"]
        }
    }')

USER_ID=$(echo $USER_RESPONSE | jq -r '.id')
echo "Created user with ID: $USER_ID"
echo "User details:"
echo $USER_RESPONSE | jq '.'
echo

# Get user by ID
echo "🔍 Getting User by ID"
echo "---------------------"
USER_BY_ID=$(curl -s ${BASE_URL}/users/${USER_ID})
echo "Retrieved user:"
echo $USER_BY_ID | jq '.'
echo

# Get user by email
echo "📧 Getting User by Email"
echo "------------------------"
USER_BY_EMAIL=$(curl -s "${BASE_URL}/users?email=demo@example.com")
echo "Retrieved user by email:"
echo $USER_BY_EMAIL | jq '.'
echo

# Create a post
echo "📝 Creating a Post"
echo "------------------"
POST_RESPONSE=$(curl -s -X POST ${BASE_URL}/posts \
    -H "Content-Type: application/json" \
    -d "{
        \"user_id\": \"$USER_ID\",
        \"title\": \"Welcome to RESTD\",
        \"content\": \"This is a demo post showcasing the RESTD microservice built with Huma framework and DDAO. It supports multiple storage engines and provides a clean REST API.\",
        \"published\": true,
        \"metadata\": {
            \"category\": \"announcement\",
            \"tags\": [\"welcome\", \"demo\", \"restd\", \"huma\", \"ddao\"],
            \"featured\": true
        }
    }")

POST_ID=$(echo $POST_RESPONSE | jq -r '.id')
echo "Created post with ID: $POST_ID"
echo "Post details:"
echo $POST_RESPONSE | jq '.'
echo

# Get post by ID
echo "📖 Getting Post by ID"
echo "---------------------"
POST_BY_ID=$(curl -s ${BASE_URL}/posts/${POST_ID})
echo "Retrieved post:"
echo $POST_BY_ID | jq '.'
echo

# Update the post
echo "✏️  Updating Post"
echo "-----------------"
UPDATED_POST=$(curl -s -X PUT ${BASE_URL}/posts/${POST_ID} \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Welcome to RESTD - Updated!",
        "content": "This is an updated demo post showcasing the RESTD microservice. The content has been modified to demonstrate the update functionality.",
        "metadata": {
            "category": "announcement",
            "tags": ["welcome", "demo", "restd", "huma", "ddao", "updated"],
            "featured": true,
            "last_editor": "demo_user"
        }
    }')

echo "Updated post:"
echo $UPDATED_POST | jq '.'
echo

# Update the user
echo "👥 Updating User"
echo "----------------"
UPDATED_USER=$(curl -s -X PUT ${BASE_URL}/users/${USER_ID} \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Demo User - Updated",
        "profile": {
            "location": "San Francisco, CA",
            "bio": "Updated demo user for RESTD - now with more details!",
            "interests": ["technology", "api", "go", "microservices"],
            "company": "DDAO Corp"
        }
    }')

echo "Updated user:"
echo $UPDATED_USER | jq '.'
echo

# Show API documentation links
echo "📚 API Documentation"
echo "--------------------"
echo "OpenAPI Documentation: ${BASE_URL}/docs"
echo "OpenAPI JSON Spec: ${BASE_URL}/openapi.json"
echo

# Show available endpoints
echo "🔗 Available Endpoints"
echo "----------------------"
echo "Health Check:"
echo "  GET  ${BASE_URL}/health"
echo
echo "Users:"
echo "  POST   ${BASE_URL}/users                - Create user"
echo "  GET    ${BASE_URL}/users/{userId}       - Get user by ID"
echo "  GET    ${BASE_URL}/users?email={email}  - Get user by email"
echo "  PUT    ${BASE_URL}/users/{userId}       - Update user"
echo "  DELETE ${BASE_URL}/users/{userId}       - Delete user"
echo
echo "Posts:"
echo "  POST   ${BASE_URL}/posts                - Create post"
echo "  GET    ${BASE_URL}/posts/{postId}       - Get post by ID"
echo "  PUT    ${BASE_URL}/posts/{postId}       - Update post"
echo "  DELETE ${BASE_URL}/posts/{postId}       - Delete post"
echo

# Test different storage engines
echo "🗄️  Storage Engine Demo"
echo "-----------------------"
echo "Current storage: SQLite (in-memory)"
echo
echo "To test other storage engines:"
echo "  PostgreSQL: ./restd --storage-engine postgres --connection-url 'postgres://user:pass@localhost/db'"
echo "  CockroachDB: ./restd --storage-engine cockroach --connection-url 'postgres://root@localhost:26257/defaultdb'"
echo "  YugabyteDB: ./restd --storage-engine yugabyte --connection-url 'postgres://yugabyte@localhost:5433/yugabyte'"
echo "  TiDB: ./restd --storage-engine tidb --connection-url 'root:@tcp(localhost:4000)/test'"
echo

# Cleanup demo (optional)
echo "🧹 Cleanup (Optional)"
echo "---------------------"
echo "To clean up the demo data:"
echo "  Delete post: curl -X DELETE ${BASE_URL}/posts/${POST_ID}"
echo "  Delete user: curl -X DELETE ${BASE_URL}/users/${USER_ID}"
echo

echo "✨ Demo completed successfully!"
echo "   Try accessing ${BASE_URL}/docs for interactive API documentation"
echo "   The server will continue running until you press Ctrl+C"

# If we started the server, wait for user input
if [ ! -z "$SERVER_PID" ]; then
    echo
    read -p "Press Enter to stop the server and exit..."
fi