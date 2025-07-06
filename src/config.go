package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AwsRegion      string `yaml:"aws_region"`
	ImageBucket    string `yaml:"image_bucket"`
	ImagePrefix    string `yaml:"image_prefix"`
	StateDir       string `yaml:"state_dir"`
	ThinPoolDevice string `yaml:"thin_pool_device"`
}

var CFG Config

func initConfig() {
	cfgFile := os.Getenv("CFG_FILE")
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}
	f, err := os.ReadFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(f, &CFG)
	if err != nil {
		log.Fatal(err)
	}
}
