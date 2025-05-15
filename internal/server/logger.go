package server

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "server", log.LstdFlags)
