package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/ClashrAuto/Clashr/config"
	C "github.com/ClashrAuto/Clashr/constant"
	"github.com/ClashrAuto/Clashr/hub"

	log "github.com/sirupsen/logrus"
)

var (
	version bool
	homedir string
)

func init() {
	flag.StringVar(&homedir, "d", "", "set configuration directory")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.Parse()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic recover")
			go main()
		}
	}()

	if version {
		fmt.Printf("Clash %s %s %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, C.BuildTime)
		return
	}

	if homedir != "" {
		if !filepath.IsAbs(homedir) {
			currentDir, _ := os.Getwd()
			homedir = filepath.Join(currentDir, homedir)
		}
		C.SetHomeDir(homedir)
	}

	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalf("Initial configuration directory error: %s", err.Error())
	}

	if err := hub.Parse(); err != nil {
		log.Fatalf("Parse config error: %s", err.Error())
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
