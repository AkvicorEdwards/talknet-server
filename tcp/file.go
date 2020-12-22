package tcp

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"talknet/def"
	"talknet/operator"
	"time"
)

type FileMapInfoTo struct {
	Ip   string
	uuid uint32
}

type FileMapInfo struct {
	Path string
	IP       string
	From     uint32
	To       []FileMapInfoTo
	CheckSum uint32
	Package  *Package
	ToMutex  sync.Mutex
	Mutex    sync.Mutex
}

var FileMap = make(map[string]*FileMapInfo)
var FileMapMutex = sync.Mutex{}

// 监听客户端向服务器发送文件的请求
func FileReceiveServer(address string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("FileReceiveServer recover", err)
		}
	}()

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("Failed to ResolveTCPAddr:", err.Error())
		return
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println("Failed to ListenTCP:", err.Error())
		return
	}

	log.Printf("FileReceiveServer listening [%s], "+
		"waiting for client to connect...\n", address)

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}
		go FileReceiveServerConnect(conn)
	}
}

// 1. 读 文件名
// 2. 写 ok
// 3. 读 文件内容
// 接收客户端发送的文件
func FileReceiveServerConnect(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("conn.Read err =", err)
		return
	}
	filename := string(buf[:n])
	fl, ok := FileMap[filename]
	if !ok {
		log.Printf("Error Prepare receive [%s]\n%v\n", filename, FileMap)
		return
	}
	if fl.IP != GetRemoteIpUseConn(conn) {
		log.Printf("Need:[%s] Got[%s]", FileMap[filename].IP,
			GetRemoteIpUseConn(conn))
		return
	}

	n, err = conn.Write([]byte("ok"))
	if err != nil {
		log.Println(err)
		return
	}

	fs, err := os.Create(def.TempDir + filename)
	if err != nil {
		return
	}
	defer func() {
		_ = fs.Close()
		FileMap[filename].Mutex.Unlock() // 文件接收成功
		log.Printf("File [%s] receive finished", filename)
		//txt, _ := ioutil.ReadFile(def.TempDir+filename)
		//log.Println("File Content", string(txt))
		if FileMap[filename].Package != nil {
			// 向接收者发送消息，代表有文件需要接受
			for k, v := range FileMap[filename].To {
				FileMap[filename].To[k].Ip = GetRemoteIpUseUUID(v.uuid)
				cli := GetConnectionUse(v.uuid)
				if cli != nil {
					//log.Println("Send to", cli.Username)
					cli.DataSend <- FileMap[filename].Package
				}
			}
		} else {
			guid := FileMap[filename].To[0].uuid
			err := os.Rename(def.TempDir+filename, def.GroupDir+filename)
			if err == nil {
				operator.AddFileInfoToGroup(guid, FileMap[filename].From, filename, filename, FileMap[filename].CheckSum)
			}
		}
	}()

	err = conn.SetDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}

	buf = make([]byte, 1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0{
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		_, _ = fs.Write(buf[:n])
	}

}

// 监听客户端请求接收文件的请求
func FileSendServer(address string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("FileSendServer recover", err)
		}
	}()

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("Failed to ResolveTCPAddr:", err.Error())
		return
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println("Failed to ListenTCP:", err.Error())
		return
	}

	log.Printf("FileSendServer listening [%s], "+
		"waiting for client to connect...\n", address)

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Println("Accept client connection exception:", err.Error())
			continue
		}
		go FileSendServerConnect(conn)
	}
}

// 1. 读 文件名
// 3. 写 文件
// 向客户端发送文件
func FileSendServerConnect(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("conn.Read err =", err)
		return
	}
	filename := string(buf[:n])
	log.Printf("Require file [%s]\n", filename)
	deleteKey := 0
	if !func() bool {
		ip := GetRemoteIpUseConn(conn)
		FileMap[filename].ToMutex.Lock()
		defer FileMap[filename].ToMutex.Unlock()
		for k, v := range FileMap[filename].To {
			if v.Ip == ip {
				deleteKey = k
				return true
			}
		}
		return false
	}() {
		log.Printf("Need:[%v] Got[%s]", FileMap[filename].To,
			GetRemoteIpUseConn(conn))
		return
	}

	path := ""
	if FileMap[filename].Path == def.GroupDir {
		path = def.GroupDir
	} else {
		path = def.TempDir
	}


	FileMap[filename].Mutex.Lock() // 等待文件接收完毕
	defer FileMap[filename].Mutex.Unlock()// 文件已发送给接收者
	fs, err := os.Open(path + filename)
	if err != nil {
		log.Println("os.Open err =", err)
		return
	}

	defer func() {
		_ = fs.Close()
		FileMap[filename].ToMutex.Lock()
		defer FileMap[filename].ToMutex.Unlock()
		for i := deleteKey; i < len(FileMap[filename].To)-1; i++ {
			FileMap[filename].To[i] = FileMap[filename].To[i+1]
		}
		FileMap[filename].To = FileMap[filename].To[:len(FileMap[filename].To)-1]
		if len(FileMap[filename].To) == 0 {
			if FileMap[filename].Path != def.GroupDir {
				err = os.Remove(path + filename)
				if err != nil {
					log.Println("Remove File", path+filename, err)
				}
			}
			FileMapMutex.Lock()
			delete(FileMap, filename)
			FileMapMutex.Unlock()
		}
	}()

	err = conn.SetDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}

	buf = make([]byte, 1024*10)
	for {
		n, err = fs.Read(buf)
		if err != nil || n == 0{
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		_, _ = conn.Write(buf[:n])
	}

}
