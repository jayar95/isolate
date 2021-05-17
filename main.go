package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/jayar95/isolate/iso"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var config string
var appConfig iso.Applications
var killme []*exec.Cmd

func main() {
	if len(os.Args) > 1 {
		config = os.Args[1]
	} else {
		log.Println("No config file passed to program. Attempting to open isolate-config.json from the current directory...")
		config = "isolate-config.json"
	}

	jsonFile, err := os.Open(config)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonBytes, &appConfig)
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	for i, application := range appConfig.Applications {
		ip := iso.MakeNetwork(application.Name, i)
		log.Printf("Network namespace created for %s on %s\n", application.Name, ip)

		log.Printf("Starting %s\n", application.Executable)
		command, err := iso.RunOnNamespace(application.Name, application.Executable)

		if err == nil {
			killme = append(killme, command)
			g := e.Group(application.Route)

			target := url.URL{}
			target.Host = ip + ":" + application.Port
			target.Scheme = application.Scheme

			g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(
				[]*middleware.ProxyTarget{
					{
						URL: &target,
					},
				},
			)))

			log.Printf(
				"Proxying requests on %s to %s for app %s\n",
				application.Route,
				ip,
				application.Name,
			)
		} else {
			log.Printf("Failed to execute %s\n", application.Executable)
			log.Println(err)
		}
	}

	go func() {
		var signal = <-signals
		log.Println(signal)

		for _, application := range appConfig.Applications {
			iso.DestroyNetwork(application.Name)
		}

		for _, command := range killme {
			err = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err != nil {
				log.Println(err)
			}
		}

		os.Exit(0)
	}()

	e.Logger.Fatal(e.Start(":80"))
}
