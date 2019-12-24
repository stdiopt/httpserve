package main

import "syscall/js"

func main() {
	js.Global().Get("document").Call("write", "hello wasm")
}
