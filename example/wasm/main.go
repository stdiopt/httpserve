package main

import (
	"log"
	"syscall/js"
	"time"

	"github.com/gohxs/prettylog"
)

func main() {
	prettylog.Global()
	js.Global().Get("document").Call("write", "hello wasm")
	log.Println("hello again")
	log.Println("hello again bigger\033[01;31mhah\033[0m")
	log.Println("Multi\nline\nstuff")
	log.Println("BBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineBig lineig line")

	for range time.NewTicker(time.Millisecond * 200).C {
		log.Println("I'm ticking:", time.Now())
	}
}
