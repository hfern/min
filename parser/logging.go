package parser

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var pause_time time.Duration

var flag_vvv *bool = flag.Bool("prs-vvv", false, "Very, Very Verbose parser logging.")
var flag_hits *bool = flag.Bool("prs-hits", false, "Log parser branches.")
var flag_logrecursion *bool = flag.Bool("prs-rec", false, "Log parser recursion.")
var flag_tokenmap *bool = flag.Bool("prs-tkmp", false, "Whitespace token map.")

func logPoolAccess(pool *tokenpool, n int, ok bool) {
	if *flag_vvv {
		if ok {
			log.Println("[CMP]TokPool: Access to ", n, "OK.")
		} else {
			log.Println("[CMP]TokPool: Access to ", n, "failed.")
		}
	}
}

func logParseRecursion(tok *State16) {
	if *flag_logrecursion || *flag_vvv {
		log.Printf("Descent: %s (%d): %d", Rul3s[tok.Rule], tok.Rule, tok.next)
	}
}

func logHitChild(pool *tokenpool, parent, child int16) {
	if *flag_hits || *flag_vvv {
		log.Printf("Child Hit: %d NEXT ( %d<%d )", pool.index, parent, child)
	}
}
func logHitElse(pool *tokenpool) {
	if *flag_hits || *flag_vvv {
		log.Println("Else Hit:", pool.index)
	}
}

func doRecursionPause() {
	if *flag_pause > 0 {
		time.Sleep(pause_time)
	}
}

func logRecursionHead(n int, d *Node) {
	if *flag_tokenmap {
		fmt.Printf("%s%d %s\n", get_n_spaces(n), n, Rul3s[d.Tok.Rule])
	}
}
