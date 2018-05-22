package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"math/rand"

	"github.com/FactomProject/factomd/Utilities/DatabaseGenerator/blockgen"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	// Seed rand incase any entrygen uses math/rand
	rand.Seed(time.Now().UnixNano())
	go StartProfiler(512*1024, true)

	var blockcount int
	var (
		loglvl     = flag.String("loglvl", "info", "Sets log level to 'debug', 'info', 'warning', or 'error'")
		configfile = flag.String("config", "", "Generator config file location.")
		genconfig  = flag.Bool("genconfig", false, "Does not run the program, but instead outputs the default config file")
	)
	flag.IntVar(&blockcount, "b", 1000, "Number of blocks to generate")

	flag.Parse()
	if *genconfig {
		c := blockgen.NewDefaultDBGeneratorConfig()
		data, err := yaml.Marshal(c)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		return
	}

	config := blockgen.NewDefaultDBGeneratorConfig()
	if *configfile != "" {
		file, err := os.OpenFile(*configfile, os.O_RDONLY, 0777)
		if err != nil {
			panic(err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(data, config)
		if err != nil {
			panic(err)
		}
		file.Close()
	}

	switch *loglvl {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		usage("Expect flag 'loglvl' to be either: debug, info, warning, or error")
	}

	//  Start time is the date of the gensis block. Each block will be 10min ahead of that
	gen, err := blockgen.NewDBGenerator(config)
	if err != nil {
		panic(err)
	}

	if config.StartTime != "" {
		start, _ := time.Parse(config.TimeFormat(), config.StartTime)

		// Need to check if the amount of blocks goes over the current timestamp
		end := time.Now().Add(time.Minute * 10 * time.Duration(blockcount))
		if !time.Now().Before(end) {
			usage(fmt.Sprintf("Too many blocks. The start time is %s, with %d blocks the end time is after the current time. The end block would be at %s", start, blockcount, end))
			return
		}
	}

	log.Infof("To run: %s", config.FactomLaunch())
	data, err := yaml.Marshal(config)
	if err != nil {
		panic(err)
	}
	log.Infof("Configuration: \n%s", string(data))
	// Creates the database
	err = gen.CreateBlocks(blockcount)
	if err != nil {
		panic(err)
	}
}

func usage(err string) {
	fmt.Println("Usage: DatabaseGenerator FLAGS")
	if err != "" {
		fmt.Println()
		fmt.Printf("%s\n", err)
	}
	os.Exit(1)
}
