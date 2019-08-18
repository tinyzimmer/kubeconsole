package main

import (
	"flag"
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/tinyzimmer/kubeconsole/pkg/k8sutils"
	"github.com/tinyzimmer/kubeconsole/pkg/term"
)

var factory k8sutils.KubernetesFactory
var controller term.Controller
var err error
var debug bool

func init() {
	factory = k8sutils.New()
	if err = factory.CreateClientSet(); err != nil {
		log.Fatalf("failed to create k8s clientset: %v", err)
	}
	flag.BoolVar(&debug, "d", false, "Stream console events to debug.log")
	flag.Parse()
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	controller = term.New(factory, debug)
	controller.Run()
}
