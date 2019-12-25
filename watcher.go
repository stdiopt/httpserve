package httpserve

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

type watcherMsg struct {
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}

// Watcher websocket handler
func (s Server) watcher(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}

	log.Println("Starting watcher for", r.RemoteAddr)
	// Start watcher
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Web socket error", err)
		return
	}

	watcher, err := fsnotify.NewWatcher() // watcher per socket
	if err != nil {
		return
	}
	defer watcher.Close()
	update := debounce(time.Millisecond*100, func() {
		err := c.WriteJSON("reload")
		if err != nil {
			log.Println("Sending msg err:", err)
		}
	})
	go func() {
		for event := range watcher.Events {
			if event.Op&fsnotify.Remove != 0 {
				continue
			}
			update()
		}
	}()

	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			log.Println("Read msg error:", err)
			break
		}
		if mt != websocket.TextMessage {
			continue
		}

		msg := watcherMsg{}
		if err = json.Unmarshal(data, &msg); err != nil {
			log.Println("Unmarshal error:", err)
			break
		}

		switch msg.Op {
		case "watch":
			var val []string
			json.Unmarshal(msg.Value, &val)
			for _, toWatch := range val {
				u, err := url.Parse(toWatch)
				if err != nil {
					log.Println("Url parse error:", err)
					return
				}
				absFile, err := filepath.Abs(u.Path[1:])
				if err != nil {
					log.Println("Filepath abs error:", err)
					return
				}
				err = watcher.Add(absFile) // remove root '/' prefix
				if err != nil {
					log.Printf("Error watching '%s (%s)' -- %s", toWatch, u.Path, err.Error())
					return
				}
			}
		case "error":
			var val []string
			json.Unmarshal(msg.Value, &val)
			log.Println("client err:", val)
		case "log":
			var val interface{}
			if err := json.Unmarshal(msg.Value, &val); err != nil {
				log.Print("log error unmarshalling:", err)
				continue
			}
			log.Println("console.log:", val)
		}
	}
}

// debounce delays the execution of fn to avoid multiple fast calls it will
// call the funcs in a routine
func debounce(d time.Duration, fn func()) func() {
	if fn == nil {
		panic("fn must be set")
	}
	controlChan := make(chan struct{})
	go func() {
		t := time.NewTimer(d)
		t.Stop()
		for {
			select {
			case <-controlChan:
				t.Reset(d)
			case <-t.C:
				go fn()
			}
		}
	}()
	return func() {
		controlChan <- struct{}{}
	}
}
