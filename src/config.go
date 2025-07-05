package main

type Config struct {
	awsRegion   string
	imageBucket string
	imagePrefix string
	stateDir    string
}

var CFG Config

func initConfig() {
	CFG = Config{
		imageBucket: "flyio-platform-hiring-challenge",
		imagePrefix: "images",
		stateDir:    "state",
		awsRegion:   "us-east-1",
	}
}
