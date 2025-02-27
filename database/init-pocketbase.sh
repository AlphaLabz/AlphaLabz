#!/bin/bash
set -e  # Exit on error

# Variables
PB_BINARY="./pocketbase"

# Function to wait for PocketBase readiness
wait_for_pocketbase() {
    local retries=30
    local wait_time=2
    local endpoint="http://localhost:8090/api/health"
    
    echo "Waiting for PocketBase to be ready..."
    while [ $retries -gt 0 ]; do
        if curl -s -f "$endpoint" > /dev/null 2>&1; then
            echo "PocketBase is ready!"
            return 0
        fi
        retries=$((retries-1))
        echo "Waiting for PocketBase to start... ($retries attempts left)"
        sleep $wait_time
    done
    echo "PocketBase did not start in time."
    return 1
}

# Start PocketBase in the background
"${PB_BINARY}" serve --http=0.0.0.0:8090 &
PB_PID=$!

# Ensure PocketBase shuts down on script exit
trap 'kill $PB_PID' EXIT

# Wait for PocketBase to be ready
wait_for_pocketbase

# Ensure there is ADMIN_EMAIL and ADMIN_PSSWORD env
if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "ADMIN_EMAIL and ADMIN_PASSWORD environment variables must be set."
    exit 1
fi

# Create or update superuser
"${PB_BINARY}" superuser upsert "${ADMIN_EMAIL}" "${ADMIN_PASSWORD}"

# Grant superuser token
RESPONSE=$(curl -s -X POST http://127.0.0.1:8090/api/collections/_superusers/auth-with-password \
-H "Content-Type: application/json" \
-d '{
    "identity": "'"${ADMIN_EMAIL}"'",
    "password": "'"${ADMIN_PASSWORD}"'"
}')
TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d '"' -f 4)

if [ -z "$TOKEN" ]; then
    echo "Failed to grant superuser token."
    exit 1
fi

echo "Superuser token granted successfully."


# User Upsert (Check if the user exists)
USER_EXISTS_RESPONSE=$(curl -s -X GET "http://127.0.0.1:8090/api/collections/users/records?filter=email='${ADMIN_EMAIL}'" \
-H "Authorization: Bearer $TOKEN")

USER_ID=$(echo "$USER_EXISTS_RESPONSE" | grep -o '"id":"[^"]*' | cut -d '"' -f 4)

# If user exists, update; otherwise, create
if [ -n "$USER_ID" ]; then
    echo "User already exists. Updating user with ADMIN role."
    USER_RESPONSE=$(curl -s -X PATCH "http://127.0.0.1:8090/api/collections/users/records/$USER_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "email": "'"${ADMIN_EMAIL}"'",
        "password": "'"${ADMIN_PASSWORD}"'",
        "passwordConfirm": "'"${ADMIN_PASSWORD}"'",
        "emailVisibility": "'"true"'",
        "verified": "'"true"'",
        "role": "'"0001"'",
        "role": "'"0001"'",
        "name": "'"admin"'",
        "user_settings": "'"0001"'"
    }')
else
    echo "Creating a new user with ADMIN role."
    USER_RESPONSE=$(curl -s -X POST "http://127.0.0.1:8090/api/collections/users/records" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "email": "'"${ADMIN_EMAIL}"'",
        "password": "'"${ADMIN_PASSWORD}"'",
        "passwordConfirm": "'"${ADMIN_PASSWORD}"'",
        "emailVisibility": "'"true"'",
        "verified": "'"true"'",
        "role": "'"0001"'",
        "name": "'"admin"'",
        "user_settings": "'"0001"'"
    }')
fi

# Confirm successful user creation
if echo "$USER_RESPONSE" | grep -q '"id":"'; then
    echo "User created or updated successfully with ADMIN role."
else
    echo "Failed to create/update user."
    echo "$USER_RESPONSE"
    exit 1
fi

echo "Setup completed successfully. PocketBase is running with roles and users created."
wait $PB_PID