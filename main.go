package main

import (
	"auth-microservice/config"
	"auth-microservice/jwt"
	userpb "auth-microservice/proto/user"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

var dbConnector *gorm.DB

type UserService struct {
	userpb.UnimplementedUserServiceServer
	jwtManager *jwt.JWTManager
}

// Responsible for starting the server
func startServer() {
	// Log a message
	fmt.Println("Starting server...")
	// Initialize the gotenv file..
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	
	// Create a new context
	dbConnector = config.ConnectDB()

	// Start the server on port 50051
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}

	// Creating a new JWT Manager
	JwtManager, err := jwt.NewJWTManager(os.Getenv("SECRET_KEY"), 5*time.Hour)
	if err != nil {
		log.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(config.UnaryInterceptor),
	)

	// Register the service with the server
	userpb.RegisterUserServiceServer(grpcServer, &UserService{jwtManager: JwtManager})

	// Start the server in a new goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %s", err)
		}
	}()

	// Create a new gRPC-Gateway server
	connection, err := grpc.DialContext(
		context.Background(),
		"localhost:50051",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	// Create a new gRPC-Gateway mux
	gwmux := runtime.NewServeMux()

	// Register the service with the gRPC-Gateway
	err = userpb.RegisterUserServiceHandler(context.Background(), gwmux, connection)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// Enable CORS
	corsOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	corsHandler := handlers.CORS(corsOrigins, corsMethods, corsHeaders)
	wrappedGwmux := corsHandler(gwmux)

	// Create a new HTTP server
	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: wrappedGwmux,
	}
	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	if err := gwServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
}

func main() {
	// Start the server
	startServer()
}
