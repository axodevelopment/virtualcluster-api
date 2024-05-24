package main

import (
	"context"
	"fmt"
	"path/filepath"

	"net/http"
	"os"

	"github.com/spf13/viper"

	sb "github.com/axodevelopment/servicebase"
	u "github.com/axodevelopment/servicebase/pkg/utils"
	"github.com/gin-gonic/gin"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	Port         int
	UKey         string
	UseLocalKube bool
}

var (
	serviceName = "VIRTUALCLUSTERS-REST"

	EnvVar    map[string]u.EnvVar
	APP_READY chan struct{}
	AppConfig *Config
)

type InvalidSetupError struct{}

func (e *InvalidSetupError) Error() string {
	return "Unable to correctly setup from configuration.  Generic message - most likely something wrong with the kubeconfig."
}

func main() {
	defer fmt.Println(serviceName + " Application Exiting ...")
	fmt.Println(serviceName + " Application Starting ...")

	var err error
	AppConfig, err = loadConfig()

	if err != nil {
		panic("Config not parsing / missing")
	}

	initSvc()

	validateSvc()

	var svc *sb.Service

	fmt.Println(serviceName + " Service.New")
	svc, _ = sb.New(serviceName, sb.WithPort(AppConfig.Port), sb.WithHealthProbe(true))

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

func getKubeClient() (*kubernetes.Clientset, error) {
	var cfg *rest.Config
	var err error

	if AppConfig.UseLocalKube {

		if homedir := homedir.HomeDir(); homedir != "" {
			p := filepath.Join(homedir, ".kube", "config")

			cfg, err = clientcmd.BuildConfigFromFlags("", p)

			if err != nil {
				return nil, err
			}

		} else {
			return nil, &InvalidSetupError{}
		}
	} else {
		cfg, err = rest.InClusterConfig()

		if err != nil {
			return nil, err
		}
	}

	clientset, cerr := kubernetes.NewForConfig(cfg)

	if cerr != nil {
		return nil, err
	}

	return clientset, nil
}

func initSvc() {
	APP_READY = make(chan struct{})
	EnvVar = make(map[string]u.EnvVar)

}

func serviceLogic(svc *sb.Service) {

	svc.GinEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, os.Args)
	})

	//for now we will test our kubeconfig create code
	client, e := getKubeClient()

	if e != nil {

		fmt.Println("Not expecting an error from getKubeClient")
		panic(e)
	}

	//test by getting pods and displaying them
	pods, err := client.CoreV1().Pods("openshift-multus").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}

	for _, v := range pods.Items {
		fmt.Println(v.Name)
	}

	close(APP_READY)
}

func startSvc(svc *sb.Service) {
	defer fmt.Println("Start Svc... Done")
	fmt.Println("Start Svc...")

	svc.AppHealthz = true
	svc.AppReadyz = true
}

func validateSvc() {
	if AppConfig.Port <= 0 {
		panic("Port should be greater then 0.")
	}
}

func loadConfig() (*Config, error) {
	viper.SetEnvPrefix("APP")

	viper.BindEnv("PORT")
	viper.BindEnv("UKEY")
	viper.BindEnv("USE_LOCAL_KUBE")

	viper.AutomaticEnv()

	config := &Config{
		Port:         viper.GetInt("PORT"),
		UKey:         viper.GetString("UKEY"),
		UseLocalKube: viper.GetBool("USE_LOCAL_KUBE"),
	}

	if config.Port <= 0 {
		fmt.Println("OsEnvVar NotFound - [APP_PORT] => defaulted to 8080")
		config.Port = 8080
	}

	fmt.Println(config)

	//for now nil error in the future validation would could prevent panic and work in a limited state ie a db connection or something
	return config, nil
}
