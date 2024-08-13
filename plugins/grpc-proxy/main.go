package main

import "github.com/luraproject/lura/v2/logging"

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger = logging.NoOp
