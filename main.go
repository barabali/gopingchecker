package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"gopkg.in/yaml.v2"
)

type ConfigFile struct {
        Service_filter string `yaml:"service_filter"`
        Check_period int `yaml:"check_period"`
		Timeout int `yaml:"timeout"`
		Config_refresh int `yaml:"config_refresh"`
}

type ConfigEnv struct {
	Service_filter string
	Check_period int
	Timeout int
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var cfg ConfigFile
    //readConfigEnv(&cfg)
	readConfigFile(&cfg)

	fmt.Println("After config read: "+cfg.Service_filter)

	//serviceNames := strings.SplitN(cfg.Service_filter, ",", -1)

	//createTargetGoroutines(clientset, serviceNames, cfg.Check_period, cfg.Timeout)
	
	go reader()
	go startMetricsServer()
	go startRESTServer()

	for {
		//get new config file
		readConfigFile(&cfg)
		serviceNames := strings.SplitN(cfg.Service_filter, ",", -1)
		refreshTargets(clientset, serviceNames, currentTargets,cfg.Check_period, cfg.Timeout)
		time.Sleep(time.Duration(cfg.Config_refresh) * time.Second)
	}
}

func createTargetGoroutines(clientset *kubernetes.Clientset, serviceNames []string, check_period int, timeout int) {
	createStartTargets(clientset, serviceNames, currentTargets)
	for _, target := range currentTargets {
		fmt.Println("Starting target Run: "+ target.Service.Name)
		go target.Run(clientset, check_period,timeout,target.Channel)
	}
}

func readConfigFile(cfg *ConfigFile) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		processError(err)
	}
}

func readConfigEnv(cfg *ConfigEnv){
	cfg.Service_filter = os.Getenv("SERVICE_FILTER")
	check_period_string := os.Getenv("CHECK_PERIOD")
	timeout_string := os.Getenv("TIMEOUT")

	var err error
	cfg.Check_period, err = strconv.Atoi(check_period_string)
	if err != nil {
		cfg.Check_period = 5
		fmt.Println("ERROR converting string to int for check period!")
	}

	err = nil
	cfg.Timeout, err = strconv.Atoi(timeout_string)
	if err != nil {
		cfg.Timeout = 5
		fmt.Println("ERROR converting string to int for timeout!")
	}
}

func processError(err error) {
    fmt.Println(err)
}


/*func (services *v1.Service) getService(name string) {
	var watched_service *v1.Service
	for _, s := range services.Items {
		if strings.HasPrefix(s.Name, "docker-go-ping") {
			watched_service = s.DeepCopy()
			fmt.Printf("Service name: %s \n", watched_service.Name)
		}
	}
}*/