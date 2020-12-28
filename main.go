package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jeromelesaux/greenserver/config"
	"github.com/jeromelesaux/greenserver/notification"
	"github.com/jeromelesaux/greenserver/persistence"
	"github.com/jeromelesaux/greenserver/web"
)

var (
	configurationFlag = flag.String("config", "", "Configuration file path.")
)

func main() {
	flag.Parse()
	/*
	*** loading configuration file ***
	 */

	if err := config.LoadConfiguration(*configurationFlag); err != nil {
		log.Fatalf("Error while loading configuration with error [%v]\n", err)
	}

	/*
	*** connect to the rds database ***
	 */
	awsRegion := config.GlobalConfiguration.AwsRegion
	dbEndpoint := config.GlobalConfiguration.DbEndpoint
	dbUser := config.GlobalConfiguration.DbUser
	dbName := config.GlobalConfiguration.DbName
	dbPassword := config.GlobalConfiguration.DbPassword
	if err := persistence.Initialise(dbEndpoint, awsRegion, dbUser, dbName, dbPassword); err != nil {
		log.Fatalf("Error while initialise the database with error (%v)\n", err)
	}

	// controller with routes definition
	// router creation
	router := gin.Default()

	// google oauth controller
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	controller := &web.Controller{}
	authorized := router.Group("/api")
	authorized.POST("/register/:token", controller.RegisterDevice)
	router.GET("/", controller.Healthy)

	// init the apple notification
	notification.Initialise()

	// start server at port 8080
	if err := router.Run(":" + config.GlobalConfiguration.Port); err != nil {
		log.Fatalf("Can not start server error :%v\n", err)
	}
}
