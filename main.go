package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jeromelesaux/greenserver/config"
	"github.com/jeromelesaux/greenserver/notification"
	"github.com/jeromelesaux/greenserver/persistence"
	"github.com/jeromelesaux/greenserver/web"
)

var (
	configurationFlag = flag.String("config", "", "Configuration file path.")
	version           = "1.1"
)

func main() {
	fmt.Fprintf(os.Stdout, "[MAIN] version : [%s]\n", version)
	flag.Parse()
	/*
	*** loading configuration file ***
	 */

	if err := config.LoadConfiguration(*configurationFlag); err != nil {
		log.Fatalf("[MAIN] Error while loading configuration with error [%v]\n", err)
	}

	// init the apple notification
	notification.Initialise()

	/*
	*** connect to the rds database ***
	 */
	awsRegion := config.GlobalConfiguration.AwsRegion
	dbEndpoint := config.GlobalConfiguration.DbEndpoint
	dbUser := config.GlobalConfiguration.DbUser
	dbName := config.GlobalConfiguration.DbName
	dbPassword := config.GlobalConfiguration.DbPassword
	if err := persistence.Initialise(dbEndpoint, awsRegion, dbUser, dbName, dbPassword); err != nil {
		log.Fatalf("[MAIN] Error while initialise the database with error (%v)\n", err)
	}

	// controller with routes definition
	// router creation
	router := gin.Default()

	// google oauth controller
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	controller := &web.Controller{}
	authorized := router.Group("/api")
	authorized.POST("/register/", controller.RegisterDevice)
	authorized.POST("/notify", controller.Notify)
	authorized.POST("/forcenotify/", controller.ForceNotify)
	router.GET("/", controller.Healthy)

	// start server at port 8080
	if err := router.Run(":" + config.GlobalConfiguration.Port); err != nil {
		log.Fatalf("[MAIN] Can not start server error :%v\n", err)
	}
}
