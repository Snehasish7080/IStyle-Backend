package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/zone/IStyle/config"
	"github.com/zone/IStyle/internal/explore"
	"github.com/zone/IStyle/internal/feed"
	"github.com/zone/IStyle/internal/middleware"
	"github.com/zone/IStyle/internal/search"
	"github.com/zone/IStyle/internal/storage"
	"github.com/zone/IStyle/internal/style"
	"github.com/zone/IStyle/internal/tag"
	"github.com/zone/IStyle/internal/user"
	"github.com/zone/IStyle/pkg/shutdown"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	env, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}

	cleanup, err := run(env)
	defer cleanup()

	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}

	// ensure the server is shutdown gracefully & app runs
	shutdown.Gracefully()
}

func run(env config.EnvVars) (func(), error) {
	app, cleanup, err := buildServer(env)
	if err != nil {
		return nil, err
	}

	// start the server
	go func() {
		app.Listen("0.0.0.0:" + env.PORT)
	}()

	// return a function to close the server and database
	return func() {
		cleanup()
		app.Shutdown()
	}, nil
}

func buildServer(env config.EnvVars) (*fiber.App, func(), error) {
	db, err := storage.BootstrapNeo4j(env.NEO4j_URI, env.NEO4jDB_NAME, env.NEO4jDB_USER, env.NEO4jDB_Password, 10*time.Second)
	if err != nil {
		return nil, nil, err
	}

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Healthy!")
	})
	// create the middleware domain
	middlewareStore := middleware.NewMiddlewareStorage(db, env.NEO4jDB_NAME)
	appMiddleware := middleware.NewAuthMiddleware(middlewareStore)

	// user domain
	userStore := user.NewUserStorage(db, env.NEO4jDB_NAME)
	userController := user.NewUserController(userStore)
	user.AddUserRoutes(app, appMiddleware, userController)

	// style domain
	styleStore := style.NewStyleStorage(db, env.NEO4jDB_NAME)
	styleController := style.NewStyleController(styleStore)
	style.AddStyleRoutes(app, appMiddleware, styleController)

	// tag domain * TODO (Relocate to separate server)
	tagStore := tag.NewTagStorage(db, env.NEO4jDB_NAME)
	tagController := tag.NewTagController(tagStore)
	tag.AddTagRoutes(app, appMiddleware, tagController)

	// feed domain
	feedStore := feed.NewFeedStorage(db, env.NEO4jDB_NAME)
	feedController := feed.NewFeedController(feedStore)
	feed.AddFeedRoutes(app, appMiddleware, feedController)

	// search domain
	searchStore := search.NewSearchStorage(db, env.NEO4jDB_NAME)
	searchController := search.NewSearchController(searchStore)
	search.AddSearchRoutes(app, appMiddleware, searchController)

	// explore domain
	exploreStore := explore.NewExploreStorage(db, env.NEO4jDB_NAME)
	exploreController := explore.NewFeedController(exploreStore)
	explore.AddExploreRoutes(app, appMiddleware, exploreController)

	return app, func() {
		storage.CloseNeo4j(db)
	}, nil
}
