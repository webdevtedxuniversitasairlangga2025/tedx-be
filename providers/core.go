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

	merchHandler "github.com/webdevtedxuniversitasairlangga/modules/merchandise/handler"
	merchRepo "github.com/webdevtedxuniversitasairlangga/modules/merchandise/repository"
	merchService "github.com/webdevtedxuniversitasairlangga/modules/merchandise/service"

	categoryHandler "github.com/webdevtedxuniversitasairlangga/modules/categories/handler"
	categoryRepo "github.com/webdevtedxuniversitasairlangga/modules/categories/repository"
	categoryService "github.com/webdevtedxuniversitasairlangga/modules/categories/service"

	ticketHandler "github.com/webdevtedxuniversitasairlangga/modules/ticket/handler"
	ticketRepo "github.com/webdevtedxuniversitasairlangga/modules/ticket/repository"
	ticketService "github.com/webdevtedxuniversitasairlangga/modules/ticket/service"

	todoHandler "github.com/webdevtedxuniversitasairlangga/modules/todo/handler"
	todoRepo "github.com/webdevtedxuniversitasairlangga/modules/todo/repository"
	todoService "github.com/webdevtedxuniversitasairlangga/modules/todo/service"

	userHandler "github.com/webdevtedxuniversitasairlangga/modules/user/handler"
	userService "github.com/webdevtedxuniversitasairlangga/modules/user/service"

	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/user/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
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
	if err := database.SeedCategories(db); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
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

	userSvc := userService.NewUserService(userRepository, db)

	do.Provide(
		injector, func(i *do.Injector) (userHandler.UserHandler, error) {
			return userHandler.NewUserHandler(i, userSvc), nil
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

	merchandiseRepository := merchRepo.NewMerchandiseRepository(db)
	merchandiseSvc := merchService.NewMerchandiseService(merchandiseRepository)

	do.Provide(
		injector, func(i *do.Injector) (merchHandler.MerchandiseHandler, error) {
			return merchHandler.NewMerchandiseHandler(i, merchandiseSvc), nil
		},
	)

	categoryRepository := categoryRepo.NewCategoryRepository(db)
	categorySvc := categoryService.NewCategoryService(categoryRepository)

	do.Provide(
		injector, func(i *do.Injector) (categoryHandler.CategoryHandler, error) {
			return categoryHandler.NewCategoryHandler(i, categorySvc), nil
		},
	)

	ticketRepository := ticketRepo.NewTicketRepository(db)
	ticketSvc := ticketService.NewTicketService(ticketRepository)

	do.Provide(
		injector, func(i *do.Injector) (ticketHandler.TicketHandler, error) {
			return ticketHandler.NewTicketHandler(i, ticketSvc), nil
		},
	)
}
