package main

import (
	"fmt"

	"net/http"
	"os"

	"github.com/spf13/viper"

	sb "github.com/axodevelopment/servicebase"
	u "github.com/axodevelopment/servicebase/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Port int
	UKey string
}

var (
	serviceName = "VIRTUALCLUSTERS-REST"

	EnvVar    map[string]u.EnvVar
	APP_READY chan struct{}
	config    *Config
)

func main() {
	defer fmt.Println(serviceName + " Application Exiting ...")
	fmt.Println(serviceName + " Application Starting ...")

	var err error
	config, err = loadConfig()

	if err != nil {
		panic("Config not parsing / missing")
	}

	initSvc()

	validateSvc()

	var svc *sb.Service

	fmt.Println(serviceName + " Service.New")
	svc, _ = sb.New(serviceName, sb.WithPort(config.Port), sb.WithHealthProbe(true))

	//TODO: May need to revisit how startSvc works this lets
	go func(svc *sb.Service) {
		//we are ready to go
		//start our dependencies
		startSvc(svc)

		fmt.Println(serviceName + " starting core logic")
		serviceLogic(svc)
	}(svc)

	//Need to wait until we are ready for the svc to go
	<-APP_READY

	//start the backend
	go func(s *sb.Service) {
		sb.Start(s)
	}(svc)

	<-svc.ExitAppChan
}

func initSvc() {
	APP_READY = make(chan struct{})
	EnvVar = make(map[string]u.EnvVar)

}

func serviceLogic(svc *sb.Service) {

	svc.GinEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, os.Args)
	})

	close(APP_READY)
}

func startSvc(svc *sb.Service) {
	defer fmt.Println("Start Svc... Done")
	fmt.Println("Start Svc...")

	svc.AppHealthz = true
	svc.AppReadyz = true
}

func validateSvc() {
	if config.Port <= 0 {
		panic("Port should be greater then 0.")
	}
}

func loadConfig() (*Config, error) {
	viper.SetEnvPrefix("APP")

	viper.BindEnv("PORT")
	viper.BindEnv("UKEY")

	viper.AutomaticEnv()

	config := &Config{
		Port: viper.GetInt("PORT"),
		UKey: viper.GetString("UKEY"),
	}

	if config.Port <= 0 {
		fmt.Println("OsEnvVar NotFound - [APP_PORT] => defaulted to 8080")
		config.Port = 8080
	}

	//for now nil error in the future validation would could prevent panic and work in a limited state ie a db connection or something
	return config, nil
}
