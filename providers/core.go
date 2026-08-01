package providers

import (
	"log"

	"github.com/webdevtedxuniversitasairlangga/config"
	"github.com/webdevtedxuniversitasairlangga/database"

	authHandler "github.com/webdevtedxuniversitasairlangga/modules/auth/handler"
	authRepo "github.com/webdevtedxuniversitasairlangga/modules/auth/repository"
	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"

	bundleHandler "github.com/webdevtedxuniversitasairlangga/modules/bundle/handler"
	bundleRepo "github.com/webdevtedxuniversitasairlangga/modules/bundle/repository"
	bundleService "github.com/webdevtedxuniversitasairlangga/modules/bundle/service"

	todoHandler "github.com/webdevtedxuniversitasairlangga/modules/todo/handler"
	todoRepo "github.com/webdevtedxuniversitasairlangga/modules/todo/repository"
	todoService "github.com/webdevtedxuniversitasairlangga/modules/todo/service"

	"github.com/webdevtedxuniversitasairlangga/modules/user/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	err := database.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := repository.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)

	authService := authService.NewAuthService(userRepository, refreshTokenRepository, jwtService, db)

	do.Provide(
		injector, func(i *do.Injector) (authHandler.AuthHandler, error) {
			return authHandler.NewAuthHandler(i, authService), nil
		},
	)

	todoRepository := todoRepo.NewTodoRepository(db)
	todoSvc := todoService.NewTodoService(todoRepository, db)

	do.Provide(
		injector, func(i *do.Injector) (todoHandler.TodoHandler, error) {
			return todoHandler.NewTodoHandler(i, todoSvc), nil
		},
	)

	bundleRepository := bundleRepo.NewBundleRepository(db)
	bundleSvc := bundleService.NewBundleService(bundleRepository, db)

	do.Provide(
		injector, func(i *do.Injector) (bundleHandler.BundleHandler, error) {
			return bundleHandler.NewBundleHandler(i, bundleSvc), nil
		},
	)
}