package user

// type registerRequest struct {
// 	Email           string `json:"email"`
// 	Password        string `json:"password"`
// 	PasswordConfirm string `json:"passwordConfirm"`
// 	Role            string `json:"role"`
// }

// Register a New User
// Only users with the "MODERATOR" or "ADMIN" role can access this API.
//
// ✅ Authorization:
// - Requires an `Authorization` header with a valid token from a MODERATOR or ADMIN.
//
// ✅ Request Body (JSON):
//
//	{
//	    "email": "test2@alphalabz.net",
//	    "password": "Test1234",
//	    "passwordConfirm": "Test1234",
//	    "role": "user" // Allowed values: "user", "moderator"
//	}
//
// ✅ Successful Response (201 Created):
//
//	{
//	    "message": "User registered successfully",
//	    "user_id": "newly-created-user-id"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Invalid JSON or missing fields
//   - 401 Unauthorized → Missing or invalid Authorization token
//   - 403 Forbidden → User is not MODERATOR or ADMIN
//   - 500 Internal Server Error → Server issue
// func HandleRegister(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient) {
// 	var registerData registerRequest

// 	// Check if the request method is POST
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get auth token
// 	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Missing or invalid Authorization token", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get request body
// 	if err := json.NewDecoder(r.Body).Decode(&registerData); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Make sure password is correct
// 	if registerData.Password != registerData.PasswordConfirm {
// 		http.Error(w, "Passwords do not match", http.StatusBadRequest)
// 		return
// 	}

// 	// Only roles without "admin" is registable
// 	role := strings.ToLower(registerData.Role)
// 	if role == "admin" {
// 		http.Error(w, "Invalid role", http.StatusBadRequest)
// 		return
// 	}

// 	// Get available roles' id and find id that matches the role
// 	availableRoles, err := pbClient.GetAvailableRoles()
// 	var roleId string
// 	if err != nil {
// 		http.Error(w, "Failed to get available roles", http.StatusInternalServerError)
// 		return
// 	} else {
// 		for _, availableRole := range availableRoles {
// 			if strings.ToLower(availableRole.Name) == role {
// 				roleId = availableRole.Id
// 				break
// 			}
// 		}
// 	}

// 	err = pbClient.NewUser(registerData.Email, registerData.Password, roleId, rawToken)
// 	if err != nil {
// 		http.Error(w, "Registration failed", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
// }
