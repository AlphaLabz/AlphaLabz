package main

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/routes/login"
	"alphalabz/pkg/routes/role"
	"alphalabz/pkg/routes/user"
	"alphalabz/pkg/settings"
	"alphalabz/pkg/smtp"
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

func setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

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

		r.Post("/oauth", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleOAuth(w, r, pbClient)
		})

		r.Post("/sso", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleSSO(w, r, pbClient)
		})
	})

	// Users route
	r.Route("/users", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			user.HandleUserList(w, r, pbClient, casbinEnforcer)
		})

		// !!! Deprecated !!!
		// r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		// 	user.HandleRegister(w, r, pbClient)
		// })

		r.Post("/invite", func(w http.ResponseWriter, r *http.Request) {
			user.HandleInviteNewUser(w, r, pbClient, casbinEnforcer, SMTPClient)
		})

		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			user.HandleSignUp(w, r, pbClient, casbinEnforcer)
		})

		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleUserRemove(w, r, pbClient)
		})

		r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleUserUpdate(w, r, pbClient)
		})
	})

	// Lab_book route
	r.Route("/lab_book", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleLabBookList(w, r, pbClient)
		})

		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleLabBookCreate(w, r, pbClient)
		})

		r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleLabBookRemove(w, r, pbClient)
		})

		r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleLabBookUpdate(w, r, pbClient)
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
		r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
			// routes.HandleScheduleUpdate(w, r, pbClient)
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

	r.Route("/roles", func(r chi.Router) {
		r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
			role.HandleRoleList(w, r, pbClient, casbinEnforcer)
		})
	})

	return r
}

func main() {
	settings, err := settings.LoadSettings("settings.yml")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Settings loaded successfully")
	}

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

	pbClient, err = pocketbase.NewPocketBase(pbHost, adminEmail, adminPassword, 10, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to initialize PocketBase client: %v", err)
	}
	pbClient.StartSuperTokenAutoRenew(adminEmail, adminPassword)

	policies, err := casbin.FetchPermissions(pbClient)
	if err != nil {
		log.Fatalf("Failed to fetch policies: %v", err)
	}

	casbinEnforcer, err = casbin.InitializeCasbin(policies)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin: %v", err)
	}

	casbinEnforcer.StartPolicyAutoReload(pbClient, 60*time.Minute)

	SMTPClient = smtp.NewSMTPClient(
		settings.Mailer.Port,
		settings.Mailer.Host,
		settings.Mailer.Username,
		settings.Mailer.Password,
		settings.Mailer.FromAddress,
		settings.Mailer.FromName)

	// Test mail
	// SMTPClient.SendMail("This is a test", "This is a test email from golang server", "lin299579@gmail.com")

	// Setup and start server
	r := setupRouter()
	port := settings.Server.Port
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
