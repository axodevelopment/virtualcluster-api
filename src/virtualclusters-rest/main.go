package main

import (
	"fmt"

	"net/http"
	"os"
	"strconv"

	sb "github.com/axodevelopment/servicebase"
	u "github.com/axodevelopment/servicebase/pkg/utils"
	"github.com/gin-gonic/gin"
)

var (
	serviceName = "VIRTUALCLUSTERS-REST"

	EnvVar    map[string]u.EnvVar
	APP_READY chan struct{}
	Port      int
	UKey      string
)

func main() {
	defer fmt.Println(serviceName + " Application Exiting ...")
	fmt.Println(serviceName + " Application Starting ...")

	initSvc()

	parseEnv()

	validateSvc()

	var svc *sb.Service

	fmt.Println(serviceName + " Service.New")
	svc, _ = sb.New("AirportApp", sb.WithPort(Port), sb.WithHealthProbe(true))

	//TODO: May need to revisit how startSvc works this lets
	go func(svc *sb.Service) {
		//we are ready to go
		//start our dependencies
		startSvc(svc)

		fmt.Println(serviceName + " starting core logic")
		serviceLogic(svc)
	}(svc)

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
	if Port <= 0 {
		panic("Port should be greater then 0.")

	}
}

// .... this doesn't seem to save much time probably should just use flag or something equivalent
func parseEnv() {
	EnvVar := u.GetEnvVars("APP_PORT", "APP_UKEY")

	//convert even if it doesn't exist sine we do this anyway
	Port = 8080

	sport := EnvVar["APP_PORT"].Value

	if sport != "" {
		iport, err := strconv.Atoi(sport)

		if err == nil {
			Port = iport
			fmt.Println("OsEnvVar Found - [APP_PORT] => set to ", Port)
		}
	} else {
		fmt.Println("OsEnvVar NotFound - [APP_PORT] => defaulted to 8080")
	}

	if !EnvVar["APP_UKEY"].Exists {
		fmt.Println("ParseEnv NotFound - [APP_UKEY] => ? applied")
		UKey = "?"
	} else {
		UKey = EnvVar["APP_UKEY"].Value
	}

}
