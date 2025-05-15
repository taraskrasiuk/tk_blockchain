package node

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "node", log.LstdFlags)
