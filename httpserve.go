//go:generate go run ./cmd/genversion -out version.go -package httpserve
//go:generate go run github.com/gohxs/folder2go -nobackup -handler assets-src assets

package httpserve

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type Server struct{}

func (s Server) Start(port int) {

	for {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Println("Err opening", port, err)
			port++
			log.Println("Trying port", port)
			continue
		}

		log.Printf("Listening at:")

		addrW := bytes.NewBuffer(nil)
		fmt.Fprintf(addrW, "    http://localhost:%d\n", port)
		addrs, err := net.InterfaceAddrs()
		for _, a := range addrs {
			astr := a.String()
			if strings.HasPrefix(astr, "192.168") ||
				strings.HasPrefix(astr, "10") {
				a := strings.Split(astr, "/")[0]
				fmt.Fprintf(addrW, "    http://%s:%d\n", a, port)
			}
		}
		log.Println(addrW.String())

		http.Serve(listener, s)
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
