package main

import (
	"go01-airbnb/config"
	"go01-airbnb/internal/middleware"
	placehttp "go01-airbnb/internal/place/delivery/http"
	placerepository "go01-airbnb/internal/place/repository"
	placeusecase "go01-airbnb/internal/place/usecase"
	"net/http"

	userhttp "go01-airbnb/internal/user/delivery/http"
	userrepository "go01-airbnb/internal/user/repository"
	userusecase "go01-airbnb/internal/user/usecase"

	"go01-airbnb/pkg/db/mysql"
	"go01-airbnb/pkg/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	env := os.Getenv("ENV")
	filename := "config/config-local.yml"
	if env == "prod" {
		filename = "config/config-prod.yml"
	}

	cfg, err := config.LoadConfig(filename)
	if err != nil {
		log.Fatalln("LoadConfig:", err)
	}

	// Khai báo DB
	db, err := mysql.NewMySQL(cfg)
	if err != nil {
		log.Fatal("Cannot connect to mysql", err)
	}

	// Khai báo hashids
	hasher := utils.NewHashIds(cfg.App.Secret, 10)

	// Khai báo các lệ thuộc
	placeRepo := placerepository.NewPlaceRepository(db)
	placeUC := placeusecase.NewPlaceUseCase(placeRepo)
	placeHdl := placehttp.NewPlaceHandler(placeUC, hasher)

	userRepo := userrepository.NewUserRepository(db)
	userUC := userusecase.NewUserUseCase(cfg, userRepo)
	userHdl := userhttp.NewUserHandler(userUC)

	mw := middleware.NewMiddlewareManager(cfg, userRepo)

	r := gin.Default()
	// Global middleware, nghĩa là tất cả các routers đều phải đi qua middleware này
	r.Use(mw.Recover())

	v1 := r.Group("/api/v1")

	v1.GET("/profiles", mw.RequiredAuth(), func(c *gin.Context) {
		user := c.MustGet("user")
		c.JSON(http.StatusOK, gin.H{"data": user})
	})

	v1.GET("/admin", mw.RequiredAuth(), mw.RequiredRoles("guest"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "OK"})
	})

	v1.GET("/places", placeHdl.GetPlaces())
	v1.GET("/places/:id", placeHdl.GetPlaceByID())
	v1.POST("/places", mw.RequiredAuth(), mw.RequiredRoles("host", "admin"), placeHdl.CreatePlace())
	v1.PUT("/places/:id", mw.RequiredAuth(), mw.RequiredRoles("host", "admin"), placeHdl.UpdatePlace())
	v1.DELETE("/places/:id", mw.RequiredAuth(), mw.RequiredRoles("host", "admin"), placeHdl.DeletePlace())

	v1.POST("/register", userHdl.Register())
	v1.POST("/login", userHdl.Login())

	r.Run(":" + cfg.App.Port)
}
