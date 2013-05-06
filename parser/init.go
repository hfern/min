package parser

import (
	"flag"
	"time"
)

var flag_pause *uint = flag.Uint("cmp-pause", 0, "Duration in milliseconds to pause in order to avoid unpaused recursion. Useful for debugging. 0 = no pause.")

func init() {
	flag.Parse()
	t := time.Duration(int(*flag_pause)) * time.Millisecond

	pause_time = time.Duration(t)
}
