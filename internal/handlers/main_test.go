package handlers

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pratikdev/url-shortener-api/internal/testutils"
)

var testPool *pgxpool.Pool
var testValidator *validator.Validate

func TestMain(m *testing.M) {
	// Global Setup
	fmt.Println("Setting up global resources...")

	// load test env vars
	err := godotenv.Load("../../.env.test")
  if err != nil {
    log.Fatal("Error loading .env.test file")
  }

	testValidator = validator.New()
	testPool = testutils.InitTestDB()
	testutils.SchemaSetup(testPool)

	// Run all unit tests in this package
	exitCode := m.Run()

	// Global Teardown
	fmt.Println("Cleaning up global resources...")
	testPool.Close()

	// Exit with the correct status code
	os.Exit(exitCode)
}