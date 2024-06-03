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
	godotenv.Load()

	// Create a new context
	dsn := config.GoDotEnvVariable("DB_CONFIG")
	dbConnector = config.ConnectDB(dsn)

	// Start the server on port 50051
	listner, err := net.Listen("tcp", "localhost:50051")
	// Check for errors
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	// Creating a new JWT Manager.
	JwtManager, _ := jwt.NewJWTManager(os.Getenv("SECRET_KEY"), 5*time.Hour)

	// Create a new gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(config.UnaryInterceptor),
	)

	// Register the service with the server
	userpb.RegisterUserServiceServer(grpcServer, &UserService{jwtManager: JwtManager})

	// Start the server in a new goroutine (concurrency) (Serve).
	// This is so that the server can continue to run while we do other things in the main function and not block the main function.
	go func() {
		if err := grpcServer.Serve(listner); err != nil {
			log.Fatalf("Failed to serve: %s", err)
		}
	}()
	// Create a new gRPC-Gateway server (gateway).
	connection, err := grpc.DialContext(
		context.Background(),
		"localhost:50051",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	// Create a new gRPC-Gateway mux (gateway).
	gwmux := runtime.NewServeMux()

	// Register the service with the server (gateway).
	err = userpb.RegisterUserServiceHandler(context.Background(), gwmux, connection)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	// Create a new HTTP server (gateway). (Serve). (ListenAndServe)
	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
func main() {
	// Start the server
	startServer()
}
