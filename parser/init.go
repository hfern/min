package parser

import (
	"flag"
	"time"
)

var flag_pause *uint = flag.Uint("prs-pause", 0, "Milliseconds for parser pausing.")

func init() {
	flag.Parse()
	t := time.Duration(int(*flag_pause)) * time.Millisecond

	pause_time = time.Duration(t)
}
