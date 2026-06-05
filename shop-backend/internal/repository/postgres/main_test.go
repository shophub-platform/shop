package postgres_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/shophub/shop/internal/model"
)

// testDB is shared across all integration tests in this package.
var testDB *gorm.DB

func TestMain(m *testing.M) {
	flag.Parse() // must be called before testing.Short()
	os.Exit(run(m))
}

func run(m *testing.M) int {
	if testing.Short() {
		// Skip container setup; individual tests will call t.Skip().
		return m.Run()
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("failed to start postgres container: %v\n", err)
		return 1
	}
	defer container.Terminate(ctx) //nolint:errcheck

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Printf("failed to get container host: %v\n", err)
		return 1
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("failed to get container port: %v\n", err)
		return 1
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=testdb sslmode=disable",
		host, port.Port(),
	)

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("failed to open gorm connection: %v\n", err)
		return 1
	}

	if err := db.AutoMigrate(
		&model.Item{},
		&model.Cart{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
	); err != nil {
		fmt.Printf("failed to migrate: %v\n", err)
		return 1
	}

	testDB = db
	return m.Run()
}
