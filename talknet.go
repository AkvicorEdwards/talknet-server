package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"talknet/def"
	"talknet/maintenance"
	"talknet/operator"
	"talknet/tcp"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Base recover", err)
		}
	}()

	if len(os.Args) >= 2 && os.Args[1] == "init" {
		maintenance.InitDatabase()
	}

	CheckDir()

	rand.Seed(time.Now().Unix())

	go maintenance.ShutDownListener()
	go TerminalConsole()

	go tcp.FileReceiveServer(def.FileReceivePort)
	go tcp.FileSendServer(def.FileSendPort)
	tcp.ListenTCP(def.MainPort)
}

func TerminalConsole() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("TerminalConsole recover", err)
		}
	}()
	time.Sleep(2 * time.Second)

	var opt int

	for {
		fmt.Println("-----------------------------")
		fmt.Println("-- 1. user list            --")
		fmt.Println("-- 2. add user             --")
		fmt.Println("-- 3. view global variable --")
		fmt.Println("-- 4. kill user connection --")
		fmt.Println("-- 5. NumGoroutine         --")
		fmt.Println("-----------------------------")

		_,_=fmt.Scanf("%d", &opt)

		//fmt.Print("\n\n\n\n\n\n\n")
		switch opt {
		case 1:
			list := operator.GetAllUser()
			for _, v := range list {
				fmt.Println("###############")
				fmt.Printf("# UUID: [%d] Username: [%s] Nickname: [%s]\n", v.Uuid, v.Username, v.Nickname)
				fmt.Printf("# Unaccepted Friends: %v\n", v.Unaccepted)
				fmt.Printf("# Friends: %v\n", v.Friends)
				fmt.Printf("# Groupss: %v\n", v.Groups)
				fmt.Println("###############")
			}
		case 2:
			var username, nickname, password string
			fmt.Println("Please input 'username' 'nickname' 'password':")
			_, _ = fmt.Scanf("%s %s %s\n", &username, &nickname, &password)
			operator.AddUser(username, nickname, password)
		case 3:
			fmt.Println("--  1. FileMap  --")
			fmt.Println("--  2. Client   --")
			_,_=fmt.Scanf("%d", &opt)
			switch opt {
			case 1:
				fmt.Println(tcp.FileMap)
				var childView string
				fmt.Println("Enter 'filename' to view info or 'exit' to exit")
				for {
					_, _ = fmt.Scanf("%s\n", &childView)
					if childView == "exit" {
						break
					}
					fmt.Printf("%s: From[%d] To[%v] IP[%s] CRC[%d]", childView, tcp.FileMap[childView].From,
						tcp.FileMap[childView].To, tcp.FileMap[childView].IP, tcp.FileMap[childView].CheckSum)
				}
			case 2:
				fmt.Println(tcp.Client)
				var childView uint32
				fmt.Println("Enter 'uuid' to view info or '0' to exit")
				for {
					_, _ = fmt.Scanf("%d\n", &childView)
					if childView == 0 {
						break
					}
					t := tcp.GetConnectionUse(childView)
					fmt.Printf("UUID:[%d] Username:[%s] Nickname:[%s]\n",
						t.UUID, t.Username, t.Nickname)
					fmt.Printf("SEQ:[%d] HeartbeatSEQ:[%d]\n", t.SEQ, t.HeartbeatSEQ)
				}
			}
		case 4:
			fmt.Println("################")
			for k := range tcp.Client {
				u, ok := operator.GetUser(k)
				if !ok {
					fmt.Println("Can not get info", k)
					continue
				}
				fmt.Printf("Username: [%s] UUID: [%d]\n", u.Username, u.Uuid)
			}
			fmt.Println("################")
			fmt.Println("Please enter the 'uuid' of the user who wants to disconnect")
			fmt.Println("Enter '0' to exit")
			var uuid uint32
			_, _ = fmt.Scanf("%d\n", &uuid)
			if uuid == 0 {
				continue
			}
			cli := tcp.GetConnectionUse(uuid)
			if cli == nil {
				continue
			}
			go func() {
				pkg := tcp.NewPackage()
				pkg.SetRequestCode(def.TerminateTheConnection)
				cli.DataSend <- &pkg
				time.Sleep(3*time.Second)
				cli.Termination <- true
			}()
		case 5:
			fmt.Printf("Number of goroutines: [%d]\n", runtime.NumGoroutine())
		}
	}
}

func CheckDir() {
	if _, err := os.Stat(def.TempDir); err != nil {
		fmt.Println("path not exists ", def.TempDir)
		err := os.MkdirAll(def.TempDir, 0711)
		if err != nil {
			log.Println("Error creating directory")
			log.Println(err)
			os.Exit(0)
		}
	}
	if _, err := os.Stat(def.GroupDir); err != nil {
		fmt.Println("path not exists ", def.GroupDir)
		err := os.MkdirAll(def.GroupDir, 0711)
		if err != nil {
			log.Println("Error creating directory")
			log.Println(err)
			os.Exit(0)
		}
	}
}