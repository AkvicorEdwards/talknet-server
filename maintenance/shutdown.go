package maintenance

import (
	"log"
	"os"
	"os/signal"
	"talknet/dam"
	"talknet/def"
	"talknet/tcp"
	"time"
)

func ShutDownListener() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ShutDownListener recover", err)
		}
	}()
	down := make(chan os.Signal, 1)
	signal.Notify(down, os.Interrupt, os.Kill)
	<-down
	go func() {
		select {
		case <-time.After(30*time.Second):
			os.Exit(0)
		}
	}()
	log.Println("Preparing to close")
	log.Println("Lock Database")
	dam.LockAll()
	log.Println("Shutdown TCP service")
	shutdownTcpService()
	log.Println("Remove received files")
	removeReceivedFiles()
	log.Println("Ready to close")
	os.Exit(0)
}

func shutdownTcpService() {
	tcp.ClientMutex.Lock()
	for _, v := range tcp.Client {
		p := tcp.NewPackage()
		p.SetRequestCode(def.TerminateTheConnection)
		v.DataSend <- &p
	}
	time.Sleep(1*time.Second)
	for _, v := range tcp.Client {
		v.Termination <- true
	}
}

func removeReceivedFiles() {
	tcp.FileMapMutex.Lock()
	for k, v := range tcp.FileMap {
		v.Mutex.Lock()
		err := os.Remove(def.TempDir + k)
		if err != nil {
			log.Println("Error: Remove file", k, err)
		}
	}
}