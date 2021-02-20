package httpserve

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

// Watcher websocket handler
func (s Server) watcher(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}

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
		err := func() error {
			mt, data, err := c.ReadMessage()
			if err != nil {
				return err
			}
			if mt != websocket.TextMessage {
				return nil
			}

			msg := []string{}
			err = json.Unmarshal(data, &msg)
			if err != nil {
				return err
			}
			/////////////
			// message handling
			/////////
			for _, toWatch := range msg {
				u, err := url.Parse(toWatch)
				if err != nil {
					return err
				}
				absFile, err := filepath.Abs(u.Path[1:])
				if err != nil {
					return err
				}
				err = watcher.Add(absFile) // remove root '/' prefix
				if err != nil {
					return fmt.Errorf("error watching '%s (%s)' -- %s", toWatch, u.Path, err.Error())
				}
			}
			return nil
		}()
		if err != nil {
			log.Println("WATCH Error:", err)
			watcher.Close()
			return
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
