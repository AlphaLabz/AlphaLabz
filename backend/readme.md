## 📌 Overview

This API provides authentication, user management, lab book handling, scheduling, and resource management.

-   **Base URL**: `<your-api-url>`
-   **Authorization**: Some routes require an `Authorization: Bearer <token>` header.

---

## 🚀 Health Check

### `GET /health`

-   ✅ **Purpose**: Check if the server is running.
-   ✅ **Response**:
    ```json
    "Server is healthy"
    ```

---

## 🔐 Authentication & Login

### `POST /login/account`

-   ✅ **Purpose**: User login via email and password.
-   ✅ **Request Body**:
    ```json
    {
        "email": "test@example.com",
        "password": "securepassword"
    }
    ```
-   ✅ **Response**:
    ```json
    {
        "message": "Login successful",
        "token": "your-auth-token"
    }
    ```
-   ❌ **Errors**:
    -   `400 Bad Request` → Missing or invalid request body.
    -   `401 Unauthorized` → Invalid credentials.

### `POST /login/oauth`

-   ✅ **Purpose**: Login via OAuth (Google, Facebook, etc.).
-   ❌ **Not implemented yet**.

### `POST /login/sso`

-   ✅ **Purpose**: Login via Single Sign-On (SSO).
-   ❌ **Not implemented yet**.

---

## 👤 User Management

### `GET /users/list`

-   ✅ **Purpose**: Fetch a list of all users.
-   ✅ **Authorization**: Requires a valid token.
-   ✅ **Response**:
    ```json
    {
        "totalUsers": 1,
        "users": [
            {
                "id": "qz73n36tig1k7z7",
                "email": "test@alphalabz.net",
                "emailVisibility": false,
                "verified": false,
                "name": "",
                "avatar": "test.png",
                "role": "ADMIN",
                "gender": "",
                "created": "2025-01-14 12:35:58.273Z",
                "updated": "2025-01-30 10:00:14.643Z"
            }
        ]
    }
    ```
-   ❌ **Errors**:
    -   `401 Unauthorized` → Missing or invalid token.

### `POST /users/register`

-   ✅ **Purpose**: Register a new user.
-   ✅ **Authorization**: Requires `MODERATOR` or `ADMIN`.
-   ✅ **Request Body**:
    ```json
    {
        "email": "test2@alphalabz.net",
        "password": "Test1234",
        "passwordConfirm": "Test1234",
        "role": "user"
    }
    ```
-   ❌ **Errors**:
    -   `401 Unauthorized` → Missing token.
    -   `403 Forbidden` → User does not have the required role.

### `DELETE /users/remove`

-   ✅ **Purpose**: Remove a user from the system.
-   ❌ **Not implemented yet**.

### `POST /users/update`

-   ✅ **Purpose**: Update user information.
-   ❌ **Not implemented yet**.

---

## 📒 Lab Book Management

### `GET /lab_book/list`

-   ✅ **Purpose**: Retrieve all lab books.
-   ❌ **Not implemented yet**.

### `POST /lab_book/create`

-   ✅ **Purpose**: Create a new lab book.
-   ❌ **Not implemented yet**.

### `DELETE /lab_book/remove`

-   ✅ **Purpose**: Delete a lab book.
-   ❌ **Not implemented yet**.

### `POST /lab_book/update`

-   ✅ **Purpose**: Update lab book details.
-   ❌ **Not implemented yet**.

---

## 📆 Schedule Management

### `GET /schedule/list`

-   ✅ **Purpose**: Retrieve all schedules.
-   ❌ **Not implemented yet**.

### `POST /schedule/create`

-   ✅ **Purpose**: Create a new schedule.
-   ❌ **Not implemented yet**.

### `DELETE /schedule/remove`

-   ✅ **Purpose**: Remove a schedule.
-   ❌ **Not implemented yet**.

### `POST /schedule/update`

-   ✅ **Purpose**: Update a schedule.
-   ❌ **Not implemented yet**.

---

## 🔗 Link Management

### `GET /link/list`

-   ✅ **Purpose**: Retrieve all links.
-   ❌ **Not implemented yet**.

### `POST /link/create`

-   ✅ **Purpose**: Create a new link.
-   ❌ **Not implemented yet**.

### `DELETE /link/remove`

-   ✅ **Purpose**: Delete a link.
-   ❌ **Not implemented yet**.

### `POST /link/update`

-   ✅ **Purpose**: Update a link.
-   ❌ **Not implemented yet**.

---

## 📂 Resources Management

### `GET /resources/list`

-   ✅ **Purpose**: Retrieve all resources.
-   ❌ **Not implemented yet**.

### `POST /resources/create`

-   ✅ **Purpose**: Create a new resource.
-   ❌ **Not implemented yet**.

### `DELETE /resources/remove`

-   ✅ **Purpose**: Remove a resource.
-   ❌ **Not implemented yet**.

### `POST /resources/update`

-   ✅ **Purpose**: Update a resource.
-   ❌ **Not implemented yet**.

---

## 🏷️ Resource Tags

### `GET /resources/tags/list`

-   ✅ **Purpose**: Retrieve all resource tags.
-   ❌ **Not implemented yet**.

### `POST /resources/tags/create`

-   ✅ **Purpose**: Create a new resource tag.
-   ❌ **Not implemented yet**.

### `DELETE /resources/tags/remove`

-   ✅ **Purpose**: Remove a resource tag.
-   ❌ **Not implemented yet**.

### `POST /resources/tags/update`

-   ✅ **Purpose**: Update a resource tag.
-   ❌ **Not implemented yet**.

---

## 📌 Notes

-   `✅ Implemented` → API is available.
-   `❌ Not implemented yet` → API is planned but not functional.

---
