package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	"betting-app-backend-go/handlers"
	"betting-app-backend-go/middleware"
	"betting-app-backend-go/services"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("[Init] ⚠️  No .env file found, using system environment variables")
	}
	
	log.Println("=== Starting Go backend server ===")

	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}

	// Initialize Firebase and MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var firebaseAuth *fbAuth.Client
	var mongoDB *mongo.Database

	// Firebase initialization
	log.Println("[Init] Attempting to initialize Firebase...")
	app, authClient, err := InitFirebase(ctx)
	if err != nil {
		log.Printf("[Init] ❌ Firebase initialization failed: %v\n", err)
		log.Println("[Init] ⚠️  Running in MOCK MODE (no real token verification)")
	} else {
		_ = app
		firebaseAuth = authClient
		log.Println("[Init] ✅ Firebase initialized successfully")
	}

	// MongoDB initialization
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("[Init] ⚠️  MONGODB_URI not set, skipping database connection")
	} else {
		mongoDBName := os.Getenv("MONGODB_DB")
		if mongoDBName == "" {
			mongoDBName = "betting"
		}
		log.Printf("[Init] Attempting to connect to MongoDB (db: %s)...\n", mongoDBName)
		client, db, err := ConnectMongo(ctx, mongoURI, mongoDBName)
		if err != nil {
			log.Printf("[Init] ❌ MongoDB connection failed: %v\n", err)
		} else {
			_ = client
			mongoDB = db
			log.Println("[Init] ✅ MongoDB connected successfully")
		}
	}

	// Setup HTTP router
	mux := http.NewServeMux()

	// Initialize services
	var walletService *services.WalletService
	var gameService *services.GameService
	var gameSettingsService *services.GameSettingsService
	var walletHandler *handlers.WalletHandler
	var gameHandler *handlers.GameHandler
	var adminHandler *handlers.AdminHandler
	var userHandler *handlers.UserHandler
	var gameSettingsHandler *handlers.GameSettingsHandler

	if mongoDB != nil {
		walletService = services.NewWalletService(mongoDB)
		gameService = services.NewGameService(mongoDB, walletService)
		gameSettingsService = services.NewGameSettingsService(mongoDB)
		walletHandler = handlers.NewWalletHandler(walletService)
		gameHandler = handlers.NewGameHandler(gameService)
		adminHandler = handlers.NewAdminHandler(walletService)
		userHandler = handlers.NewUserHandler(mongoDB)
		gameSettingsHandler = handlers.NewGameSettingsHandler(gameSettingsService)
		
		// Initialize default game settings if they don't exist
		if err := gameSettingsService.InitializeDefaultSettings(context.Background()); err != nil {
			log.Printf("[Init] ⚠️ Failed to initialize default game settings: %v\n", err)
		} else {
			log.Println("[Init] ✅ Game settings initialized")
		}
		
		log.Println("[Init] ✅ All services and handlers initialized")
	}

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(firebaseAuth)

	// Health check endpoint (no auth required)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":        "running",
			"firebaseReady": firebaseAuth != nil,
			"mongodbReady":  mongoDB != nil,
			"timestamp":     time.Now().Unix(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Auth verify endpoint (no auth middleware, handles its own verification)
	mux.HandleFunc("/api/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[AuthVerify] %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			IdToken string `json:"idToken"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("[AuthVerify] ❌ Invalid JSON body: %v\n", err)
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		if body.IdToken == "" {
			log.Println("[AuthVerify] ❌ Missing idToken in request")
			http.Error(w, "idToken required", http.StatusBadRequest)
			return
		}

		log.Printf("[AuthVerify] Received idToken (length: %d)\n", len(body.IdToken))

		// If Firebase auth client is available, verify token
		if firebaseAuth != nil {
			log.Println("[AuthVerify] Verifying token with Firebase...")
			tok, err := firebaseAuth.VerifyIDToken(context.Background(), body.IdToken)
			if err != nil {
				log.Printf("[AuthVerify] ❌ Token verification failed: %v\n", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			uid := tok.UID
			email := ""
			name := ""
			phone := ""

			if v, ok := tok.Claims["email"]; ok {
				if s, ok := v.(string); ok {
					email = s
				}
			}
			if v, ok := tok.Claims["name"]; ok {
				if s, ok := v.(string); ok {
					name = s
				}
			}
			if v, ok := tok.Claims["phone_number"]; ok {
				if s, ok := v.(string); ok {
					phone = s
				}
			}

			// Use phone as name fallback
			if name == "" && phone != "" {
				name = phone
			}

			log.Printf("[AuthVerify] ✅ Token verified successfully - uid: %s, phone: %s\n", uid, phone)

			// Upsert into MongoDB if available
			if mongoDB != nil {
				log.Printf("[AuthVerify] Upserting user to MongoDB (uid: %s)...\n", uid)
				if err := UpsertUser(context.Background(), mongoDB, uid, email, name); err != nil {
					log.Printf("[AuthVerify] ❌ MongoDB upsert failed: %v\n", err)
				} else {
					log.Println("[AuthVerify] ✅ User upserted successfully")
				}
			}

			response := map[string]interface{}{
				"uid":   uid,
				"email": email,
				"name":  name,
			}
			log.Printf("[AuthVerify] ✅ Sent response: %v\n", response)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Fallback mock response when Firebase is not available
		log.Println("[AuthVerify] ⚠️  Firebase not available, returning mock response")
		mockResponse := map[string]interface{}{
			"uid":   "mock-uid-123",
			"email": "user+mock@example.com",
			"name":  "Mock User",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	// Protected wallet endpoints
	if walletHandler != nil {
		mux.Handle("/api/wallet/balance", authMiddleware(http.HandlerFunc(walletHandler.GetBalance)))
		mux.Handle("/api/wallet/payment-request", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				walletHandler.CreatePaymentRequest(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		mux.Handle("/api/wallet/payment-requests", authMiddleware(http.HandlerFunc(walletHandler.GetPaymentRequests)))
		mux.Handle("/api/wallet/transactions", authMiddleware(http.HandlerFunc(walletHandler.GetTransactions)))
		mux.HandleFunc("/api/wallet/payment-details", walletHandler.GetPaymentDetails)
		log.Println("[Init] ✅ Wallet endpoints registered")
	}

	// Protected game endpoints
	if gameHandler != nil {
		mux.Handle("/api/game/play", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				gameHandler.PlayGame(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		mux.Handle("/api/game/history", authMiddleware(http.HandlerFunc(gameHandler.GetGameHistory)))
		mux.Handle("/api/game/stats", authMiddleware(http.HandlerFunc(gameHandler.GetGameStats)))
		mux.HandleFunc("/api/game/recent-bets", gameHandler.GetRecentBets)
		log.Println("[Init] ✅ Game endpoints registered")
	}

	// Protected admin endpoints (requires admin role)
	if adminHandler != nil {
		mux.Handle("/api/admin/payment-requests", authMiddleware(middleware.RequireAdmin(http.HandlerFunc(adminHandler.GetAllPaymentRequests))))
		mux.Handle("/api/admin/payment-request/", authMiddleware(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				adminHandler.ProcessPaymentRequest(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}))))
		mux.Handle("/api/admin/payment-details", authMiddleware(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
				adminHandler.UpdatePaymentDetails(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}))))
		log.Println("[Init] ✅ Admin endpoints registered")
	}

	// Protected user endpoints
	if userHandler != nil {
		mux.Handle("/api/user/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				userHandler.GetProfile(w, r)
			} else if r.Method == http.MethodPut {
				userHandler.UpdateProfile(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		mux.Handle("/api/user/stats", authMiddleware(http.HandlerFunc(userHandler.GetStats)))
		log.Println("[Init] ✅ User endpoints registered")
	}

	// Game settings endpoints (public read, admin write)
	if gameSettingsHandler != nil {
		mux.HandleFunc("/api/game-settings", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				gameSettingsHandler.GetAllGameSettings(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})
		mux.Handle("/api/game-settings/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				gameSettingsHandler.GetGameSettings(w, r)
			} else if r.Method == http.MethodPut {
				// Require auth and admin for updates
				authMiddleware(middleware.RequireAdmin(http.HandlerFunc(gameSettingsHandler.UpdateGameSettings))).ServeHTTP(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}))
		log.Println("[Init] ✅ Game settings endpoints registered")
	}

	// CORS middleware
	withCORS := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
			if allowedOrigin == "" {
				allowedOrigin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Create and start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: withCORS(mux),
	}

	log.Printf("=== Go backend running on http://localhost:%s ===\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
