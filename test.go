package main

import (
	"fmt"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create, watcher.Write)

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event.Path)
				fmt.Println(event)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.Add("/data/logs"); err != nil {
		log.Fatalln(err)
	}

	// go func() {
	// 	w.Wait()
	// 	w.TriggerEvent(watcher.Create, nil)
	// 	w.TriggerEvent(watcher.Write, nil)
	// }()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
