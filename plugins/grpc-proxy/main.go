package main

import "github.com/luraproject/lura/logging"

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger = logging.NoOp
