package main

import (
	"flag"
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/tinyzimmer/kubeconsole/pkg/k8sutils"
	"github.com/tinyzimmer/kubeconsole/pkg/server"
	"github.com/tinyzimmer/kubeconsole/pkg/term"
)

var factory k8sutils.KubernetesFactory
var controller term.Controller
var err error

// flags
var debug bool
var listen bool
var incluster bool
var serverkey string

func init() {

	flag.BoolVar(&debug, "d", false, "Stream console events to debug.log")
	flag.BoolVar(&listen, "listen", false, "Start SSH Server for remote connections")
	flag.StringVar(&serverkey, "keyfile", "", `A pre-generated server key file for SSH
		If you do not supply this, one will be generated`)
	flag.BoolVar(&incluster, "cluster", false, "Use in-cluster k8s config")
	flag.Parse()

}

func main() {

	if listen {
		s, err := server.New(incluster, serverkey)
		if err != nil {
			log.Fatal(err)
		}
		if err = s.Listen(); err != nil {
			log.Fatal(err)
		}
		return
	}

	factory = k8sutils.New(incluster)
	if err = factory.CreateClientSet(); err != nil {
		log.Fatalf("failed to create k8s clientset: %v", err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	controller = term.New(factory, debug)
	if err = controller.Run(); err != nil {
		log.Fatal(err)
	}

}
