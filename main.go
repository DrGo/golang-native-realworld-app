package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof" // for profiling
	"os"
	"runtime/pprof"

	"github.com/drgo/realworld/errors"
	// _ "github.com/ianlancetaylor/cgosymbolizer" 	//does not work on macOS
)

const (
	host        = "localhost"
	port        = "8080"
	apiVersion  = "v1"
	apiRoot     = "api"
	cookieName  = "session"
	maxLifeTime = 10 * 60 //10mins

	versionMessage = "%s %s (%s). CopyRight 2018-2021 Salah Mahmud"
)

var (
	exeName string = os.Args[0]
	//Version holds the exe version initialized in the Makefile
	version string
	//Build holds the exe build number initialized in the Makefile
	build string
)

func getVersion() string {
	return fmt.Sprintf(versionMessage, exeName, version, build)
}

var (
	debug      = flag.Bool("debug", false, "turn on debugging mode")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
)

func main() {
	flag.Parse()
	fmt.Println(getVersion())
	errors.Debug = *debug
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			errors.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	opts := ServerOptions{
		CookieName:   cookieName,
		MaxLifeTime:  maxLifeTime,
		DatabaseName: "db/rw.db",
		// use "localhost:8080" to suppress macos firewall permission
		Addr: host + ":" + port,
	}
	s := NewServer(&opts)
	defer s.Finalize()
	fmt.Println("debug:", errors.Debug)
	s.Start()
}
