package main

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/routes/labbook"
	"alphalabz/pkg/routes/login"
	"alphalabz/pkg/routes/role"
	"alphalabz/pkg/routes/user"
	"alphalabz/pkg/settings"
	"alphalabz/pkg/smtp"
	"alphalabz/pkg/tools"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

var pbClient *pocketbase.PocketBaseClient
var casbinEnforcer *casbin.CasbinEnforcer
var SMTPClient *smtp.SMTPClient

// JWTAuthMiddleware will check the JWT token and validate it.
// If valid, it will pass the request to the next handler. Otherwise, it will return a 401 Unauthorized response.
func JWTExpirationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtSkipPaths = map[string]bool{
			"/health":        true,
			"/login/account": true,
			// "/login/oauth":   true,
			// "/login/sso":     true,
			"/users/signup": true,
		}

		// Check if the path is in the skip list. If it is, then skip JWT validation and pass the request to the next handler.
		if _, ok := jwtSkipPaths[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "token missing or invalid", http.StatusUnauthorized)
			return
		}

		// Verify JWT expiration. If the token has expired or is invalid, return a 401 Unauthorized response.
		valid, err := tools.VerifyJWTExpiration(rawToken)
		if err != nil || !valid {
			http.Error(w, "token expired or invalid", http.StatusUnauthorized)
			return
		}

		// Token is valid. Pass the request to the next handler.
		next.ServeHTTP(w, r)
	})
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Use(JWTExpirationMiddleware)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Server is healthy")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Corrected JSON encoding
		response := map[string]string{"message": "server is healthy"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Login to system
	r.Route("/login", func(r chi.Router) {
		r.Post("/account", func(w http.ResponseWriter, r *http.Request) {
			login.HandleAccountLogin(w, r, pbClient)
		})

		// r.Post("/oauth", func(w http.ResponseWriter, r *http.Request) {
		// 	// routes.HandleOAuth(w, r, pbClient)
		// })

		// r.Post("/sso", func(w http.ResponseWriter, r *http.Request) {
		// 	// routes.HandleSSO(w, r, pbClient)
		// })
	})

	// Users route
	r.Route("/user", func(r chi.Router) {
		r.Get("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
			userId := chi.URLParam(r, "id")
			user.HandleUserView(w, r, userId, pbClient, casbinEnforcer)
		})

		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			user.HandleUserList(w, r, pbClient, casbinEnforcer)
		})

		r.Post("/invite", func(w http.ResponseWriter, r *http.Request) {
			user.HandleInviteNewUser(w, r, pbClient, casbinEnforcer, SMTPClient)
		})

		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			user.HandleSignUp(w, r, pbClient, casbinEnforcer)
		})

		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			user.HandleUserRemove(w, r, pbClient, casbinEnforcer)
		})

		r.Patch("/settings", func(w http.ResponseWriter, r *http.Request) {
			user.HandlUpdateSettings(w, r, pbClient, casbinEnforcer)
		})

		r.Route("/account", func(r chi.Router) {
			// r.Patch("/modify/email", func(w http.ResponseWriter, r *http.Request) {})
			// r.Patch("/modify/password", func(w http.ResponseWriter, r *http.Request) {})

			// for name, birthdate, gender
			r.Patch("/update", func(w http.ResponseWriter, r *http.Request) {
				user.HandlUpdateProfile(w, r, pbClient, casbinEnforcer)
			})

			// for avatar only
			r.Patch("/update/avatar", func(w http.ResponseWriter, r *http.Request) {
				user.HandleUpdateAvatar(w, r, pbClient, casbinEnforcer)
			})
		})
	})

	// Lab_book route
	r.Route("/labbook", func(r chi.Router) {
		r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			labbook.HandleLabBookUpload(w, r, pbClient, casbinEnforcer)
		})

		r.Get("/upload/history", func(w http.ResponseWriter, r *http.Request) {
			labbook.HandleLabbookUploadHistory(w, r, pbClient, casbinEnforcer)
		})

		r.Post("/share", func(w http.ResponseWriter, r *http.Request) {
			labbook.HandleShareLabbook(w, r, pbClient, casbinEnforcer)
		})
		r.Get("/shared/list", func(w http.ResponseWriter, r *http.Request) {
			labbook.GetSharedList(w, r, pbClient, casbinEnforcer)
		})

		r.Get("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
			labbookId := chi.URLParam(r, "id")
			labbook.HandleLabBookView(w, r, labbookId, pbClient, casbinEnforcer)
		})

		// r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
		// routes.HandleLabBookRemove(w, r, pbClient)
		// })

		// r.Patch("/update", func(w http.ResponseWriter, r *http.Request) {
		// routes.HandleLabBookUpdate(w, r, pbClient)
		// })

		r.Patch("/review", func(w http.ResponseWriter, r *http.Request) {
			labbook.HandleLabBookReview(w, r, pbClient, casbinEnforcer)
		})

		r.Get("/review/pending", func(w http.ResponseWriter, r *http.Request) {
			labbook.GetPendingReviews(w, r, pbClient, casbinEnforcer)
		})

		r.Get("/reviewers", func(w http.ResponseWriter, r *http.Request) {
			labbook.GetAvailiableReviewers(w, r, pbClient, casbinEnforcer)
		})
	})

	r.Route("/roles", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			role.HandleRoleList(w, r, pbClient, casbinEnforcer)
		})

		r.Get("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
			// roleId := chi.URLParam(r, "id")
			// role.HandleRoleView(w, r, roleId, pbClient, casbinEnforcer)
		})

		// r.Patch("/update", func(w http.ResponseWriter, r *http.Request) {})

		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			role.HandleCreateNewRole(w, r, pbClient, casbinEnforcer)
		})

		r.Delete("/remove/{id}", func(w http.ResponseWriter, r *http.Request) {
			deleteRoleId := chi.URLParam(r, "id")
			role.HandleDeleteRole(w, r, deleteRoleId, pbClient, casbinEnforcer)
		})
	})

	// Schedule routes
	r.Route("/schedule", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleList(w, r, pbClient)
		})
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleCreate(w, r, pbClient)
		})
		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleRemove(w, r, pbClient)
		})
		r.Patch("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleUpdate(w, r, pbClient)
		})

		r.Patch("/share", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleShare(w, r, pbClient)
		})
	})

	// Link routes
	r.Route("/link", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleList(w, r, pbClient)
		})
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleCreate(w, r, pbClient)
		})
		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleRemove(w, r, pbClient)
		})
		r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleUpdate(w, r, pbClient)
		})
	})

	r.Route("/resources", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleList(w, r, pbClient)
		})
		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleCreate(w, r, pbClient)
		})
		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleRemove(w, r, pbClient)
		})
		r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleUpdate(w, r, pbClient)
		})
	})

	r.Route("/system", func(r chi.Router) {
		r.Patch("/settings", func(w http.ResponseWriter, r *http.Request) {})

		r.Patch("/settings/smtp", func(w http.ResponseWriter, r *http.Request) {})

		r.Route("/plugin", func(r chi.Router) {
			r.Post("/install", func(w http.ResponseWriter, r *http.Request) {})
		})
	})

	return r
}

func main() {
	// Initialize settings from YAML file
	settings, err := settings.LoadSettings("settings.yml")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Settings loaded successfully")
	}

	// Initialize SMTP client
	SMTPClient = smtp.NewSMTPClient(
		settings.Mailer.Port,
		settings.Mailer.Host,
		settings.Mailer.Username,
		settings.Mailer.Password,
		settings.Mailer.FromAddress,
		settings.Mailer.FromName,
	)

	// Get PocketBase host from env or default to local development server
	pbHost := os.Getenv("POCKETBASE_URL")
	if pbHost == "" {
		pbHost = "http://127.0.0.1:8090"
	}

	// Get admin password from env
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" || adminPassword == "" {
		log.Fatal("Missing required environment variables: ADMIN_EMAIL and ADMIN_PASSWORD")
	}

	// Initialize PocketBase client with admin credentials and start supertoken auto-renewal
	pbClient, err = pocketbase.NewPocketBase(pbHost, adminEmail, adminPassword, 10, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to initialize PocketBase client: %v", err)
	}
	pbClient.StartSuperTokenAutoRenew(adminEmail, adminPassword)

	policies, err := casbin.FetchPermissions(pbClient)
	if err != nil {
		log.Fatalf("Failed to fetch policies: %v", err)
	}

	// Initialize Casbin with policies
	casbinEnforcer, err = casbin.InitializeCasbin(policies)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin: %v", err)
	}

	// Start policy auto-reload every 60 minutes
	casbinEnforcer.StartPolicyAutoReload(pbClient, 60*time.Minute)

	// Create uploads directory if it doesn't exist
	if err = tools.CreateUploadsDir(); err != nil {
		log.Fatal("Failed to create uploads directory")
	}

	// Start auto clean uploads every 24 hours
	tools.StartAutoCleanUploads(24 * time.Hour)

	// Setup and start server
	r := setupRouter()
	port := settings.Server.Port
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
