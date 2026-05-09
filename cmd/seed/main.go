// Command seed creates the first church_admin user when none exists.
// Run it once to bootstrap admin access; afterwards use POST /admin/user/create.
//
// Usage:
//
//	go run ./cmd/seed \
//	  -email admin@hofng.org \
//	  -first-name John \
//	  -last-name Doe \
//	  -password 'SecretPass123!'
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/database"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/logger"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/persistence"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	"go.uber.org/zap"
)

func main() {
	email := flag.String("email", "", "Admin email address (required)")
	firstName := flag.String("first-name", "", "Admin first name (required)")
	lastName := flag.String("last-name", "", "Admin last name (required)")
	password := flag.String("password", "", "Admin password — min 8 chars (required)")
	flag.Parse()

	if *email == "" || *firstName == "" || *lastName == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/seed -email <email> -first-name <name> -last-name <name> -password <pass>")
		os.Exit(1)
	}
	if len(*password) < 8 {
		fmt.Fprintln(os.Stderr, "error: password must be at least 8 characters")
		os.Exit(1)
	}

	_ = godotenv.Overload()

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	zapLog, err := logger.New(appEnv)
	if err != nil {
		log.Fatalf("initializing logger: %v", err)
	}

	cfg, err := config.Load(zapLog)
	if err != nil {
		zapLog.Fatal("loading config", zap.Error(err))
	}

	db, err := database.Connect(&cfg.Database, zapLog)
	if err != nil {
		zapLog.Fatal("connecting to database", zap.Error(err))
	}

	if migrateErr := database.RunMigrations(db, "./migrations", zapLog); migrateErr != nil {
		zapLog.Fatal("running migrations", zap.Error(migrateErr))
	}

	ctx := context.Background()
	repo := persistence.NewUserRepository(db, zapLog)

	// Option C guard: refuse if any church_admin already exists.
	var count int64
	if countErr := db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", domainUser.RoleChurchAdmin).
		Count(&count).Error; countErr != nil {
		zapLog.Fatal("checking existing admins", zap.Error(countErr))
	}
	if count > 0 {
		zapLog.Fatal("admin already exists — use POST /admin/user/create (requires admin JWT) to add more")
	}

	hashed, err := security.HashPassword(*password)
	if err != nil {
		zapLog.Fatal("hashing password", zap.Error(err))
	}

	u := &domainUser.User{
		FirstName:       *firstName,
		LastName:        *lastName,
		Email:           *email,
		Password:        hashed,
		PasswordVersion: domainUser.PasswordVersionBcrypt,
		IsVerified:      domainUser.EmailVerified,
	}

	if err := repo.Create(ctx, u); err != nil {
		zapLog.Fatal("creating admin user", zap.Error(err))
	}

	if err := repo.AssignRoles(ctx, u.ID, []domainUser.RoleName{domainUser.RoleChurchAdmin}); err != nil {
		zapLog.Fatal("assigning church_admin role", zap.Error(err))
	}

	fmt.Printf("✓ Admin created: %s %s <%s>\n", *firstName, *lastName, *email)
	fmt.Println("  Sign in at POST /session/sign_in/admin")
}
