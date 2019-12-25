package main

import "syscall/js"

func main() {
	js.Global().Get("document").Call("write", "hello wasm")
	js.Global().Get("console").Call("log", "hello log")
}
