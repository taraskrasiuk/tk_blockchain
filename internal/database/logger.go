package database

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "database.", log.LstdFlags)
