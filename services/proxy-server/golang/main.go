package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

const (
	defaultPort    = 8001
	baseBucketPath = "turbo-deploy-s3"
	region         = "ap-south-1"
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	idleTimeout    = 120 * time.Second
	maxHeaderBytes = 1 << 20 // 1 MB
)

type ReverseProxy struct {
	limiter    *rate.Limiter
	s3Client   *s3.Client
	bucketName string
}

func NewReverseProxy() *ReverseProxy {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	return &ReverseProxy{
		limiter:    rate.NewLimiter(rate.Limit(100), 200),
		s3Client:   s3Client,
		bucketName: baseBucketPath,
	}
}

func (rp *ReverseProxy) validateAndGetDeployment(subdomain string) (string, error) {
	// Validate subdomain format
	validSubdomainRegex := regexp.MustCompile(`^[a-zA-Z0-9-_]{1,63}$`)
	if !validSubdomainRegex.MatchString(subdomain) {
		return "", fmt.Errorf("invalid subdomain format")
	}

	// Construct the full path to check
	deploymentPath := fmt.Sprintf("__outputs/%s/index.html", subdomain)

	// Check if the deployment exists in S3
	_, err := rp.s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(rp.bucketName),
		Key:    aws.String(deploymentPath),
	})

	if err != nil {
		log.Printf("Deployment not found for subdomain %s: %v", subdomain, err)
		return "", fmt.Errorf("deployment not found")
	}

	// Use the correct S3 endpoint for the region
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/__outputs/%s", 
		rp.bucketName, region, subdomain), nil
}

func (rp *ReverseProxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if err := rp.limiter.Wait(r.Context()); err != nil {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Extract subdomain
	hostname := r.Host
	parts := strings.Split(hostname, ".")

	// Handle localhost and other development environments
	var subdomain string
	if len(parts) > 1 {
		subdomain = parts[0]
	}

	// Validate and get deployment URL
	targetURLStr, err := rp.validateAndGetDeployment(subdomain)
	if err != nil {
		log.Printf("Deployment validation failed: %v", err)
		http.Error(w, "Invalid or Not Found Deployment", http.StatusNotFound)
		return
	}

	// Parse target URL
	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		log.Printf("Error parsing target URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Modify request path
	originalPath := r.URL.Path
	if originalPath == "/" {
		r.URL.Path = "/index.html"
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Advanced error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for subdomain %s: %v", subdomain, err)

		switch {
		case strings.Contains(err.Error(), "PermanentRedirect"):
			http.Error(w, "S3 Endpoint Configuration Error", http.StatusInternalServerError)
		case strings.Contains(err.Error(), "connect: connection refused"):
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		case strings.Contains(err.Error(), "no such host"):
			http.Error(w, "Destination Not Found", http.StatusNotFound)
		default:
			http.Error(w, "Proxy Error", http.StatusBadGateway)
		}
	}

	// Log proxy details for debugging
	log.Printf("Proxying request: Subdomain=%s, Original Path=%s, Target URL=%s",
		subdomain, originalPath, targetURLStr)

	// Serve the proxy request
	proxy.ServeHTTP(w, r)
}

func main() {
	// Get port from environment or use default
	port := defaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		if parsedPort, err := strconv.Atoi(envPort); err == nil {
			port = parsedPort
		}
	}

	// Create reverse proxy handler
	rp := NewReverseProxy()

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server with timeouts and size limits
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        corsHandler.Handler(http.HandlerFunc(rp.handleProxy)),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Reverse Proxy running on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Block until shutdown signal
	<-stop

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}