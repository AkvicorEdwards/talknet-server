package tcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"talknet/def"
	"time"
)

func UInt16ToBytes(i uint16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, i)
	return buf
}

func BytesToUInt16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func UInt32ToBytes(i uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, i)
	return buf
}

func BytesToUInt32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

func UInt64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func BytesToUInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func CRC32(str []byte) uint32 {
	return crc32.ChecksumIEEE(str)
}

// 解包用户的账号密码
func UnwrapLoginData(data []byte) (username, password string) {
	if len(data) != LengthHeadData {
		return "", ""
	}
	return string(data[def.LoginDataOffsetUsername :def.LoginDataOffsetUsername+data[def.LoginDataOffsetUsernameLength]]),
		string(data[def.LoginDataOffsetPassword :def.LoginDataOffsetPassword+data[def.LoginDataOffsetPasswordLength]])
}

// 将用户的信息打包
func WrapUserInfo(uuid uint32, username, nickname string) []byte {
	if len([]byte(username)) > def.UserInfoLengthUsername || len([]byte(nickname)) > def.UserInfoLengthNickname {
		return []byte{}
	}
	data := make([]byte, LengthHeadData)
	id := UInt32ToBytes(uuid)
	user := []byte(username)
	nick := []byte(nickname)
	data[def.UserInfoOffsetUsernameLength] = byte(len(user))
	data[def.UserInfoOffsetNicknameLength] = byte(len(nick))
	for k, v := range id {
		data[def.UserInfoOffsetUUID+k] = v
	}
	for k, v := range user {
		data[def.UserInfoOffsetUsername+k] = v
	}
	for k, v := range nick {
		data[def.UserInfoOffsetNickname+k] = v
	}
	return data
}

// 打包短信息
func WrapMessage(uuid uint32, message string) ([]byte, bool)  {
	if len([]byte(message)) > def.MessageLengthMessage {
		return []byte{}, false
	}
	data := make([]byte, LengthHeadData)
	user := UInt32ToBytes(uuid)
	mess := []byte(message)
	data[def.MessageOffsetMessageLength] = byte(len(mess))
	for k, v := range user {
		data[def.MessageOffsetUUID+k] = v
	}
	for k, v := range mess {
		data[def.MessageOffsetMessage+k] = v
	}
	return data, true
}

// 打包短信息
func WrapGroupMessage(guid, uuid uint32, message string) ([]byte, bool)  {
	if len([]byte(message)) > def.MessageLengthGroupMessage {
		return []byte{}, false
	}
	data := make([]byte, LengthHeadData)
	guidd := UInt32ToBytes(guid)
	uuidd := UInt32ToBytes(uuid)
	mess := []byte(message)
	data[def.MessageOffsetGroupMessageLength] = byte(len(mess))
	for k, v := range guidd {
		data[def.MessageOffsetGroupGUID+k] = v
	}
	for k, v := range uuidd {
		data[def.MessageOffsetGroupUUID+k] = v
	}
	for k, v := range mess {
		data[def.MessageOffsetMessage+k] = v
	}

	return data, true
}

func WrapGuidUuid(guid, uuid uint32) []byte {
	data := UInt32ToBytes(guid)
	data = append(data, UInt32ToBytes(uuid)...)
	return data
}

func UnwrapGuidUuid(data []byte) (guid, uuid uint32) {
	guid = BytesToUInt32(data[:4])
	uuid = BytesToUInt32(data[4:8])
	return
}

func RandomFilename() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
}

// 解包消息数据
func UnwrapMessage(data []byte) (uuid uint32, message string) {
	if len(data) != LengthHeadData {
		return 0, ""
	}
	return BytesToUInt32(data[def.MessageOffsetUUID:def.MessageOffsetUUID+def.MessageLengthUUID]),
		string(data[def.MessageOffsetMessage:def.MessageOffsetMessage+data[def.MessageOffsetMessageLength]])
}

func UnwrapGroupMessage(data []byte) (guid, uuid uint32, message string) {
	if len(data) != LengthHeadData {
		return 0, 0, ""
	}
	return BytesToUInt32(data[def.MessageOffsetGroupGUID:def.MessageOffsetGroupGUID+def.MessageLengthGroupGUID]),
		BytesToUInt32(data[def.MessageOffsetGroupUUID:def.MessageOffsetGroupUUID+def.MessageLengthGroupUUID]),
		string(data[def.MessageOffsetGroupMessage:def.MessageOffsetGroupMessage+data[def.MessageOffsetGroupMessageLength]])
}

// 将头数据中的 '接收者uuid' 替换为 '发送者uuid'
func ReWrapHeadDataMessageUUID(from uint32, data *Package) {
	ori := data.GetHeadData()
	user := UInt32ToBytes(from)
	for k, v := range user {
		ori[def.MessageOffsetUUID+k] = v
	}
}

func FileReceivePrepare(filename, ip string, from uint32, to []FileMapInfoTo, checkSum uint32) *FileMapInfo {
	// 将信息打包以filename为key存储在 FileMap 中
	info := &FileMapInfo{
		IP:       ip,
		From:     from,
		To:       to,
		CheckSum: checkSum,
		Mutex: sync.Mutex{},
	}

	info.Mutex.Lock() // 文件还未接收

	FileMapMutex.Lock()
	FileMap[filename] = info
	FileMapMutex.Unlock()
	//log.Println("Add", filename, "to FileMap")
	return info
}

func AutoAdaptSendMessage(requestCode uint16, cli *Connection, uuid uint32, message string, ack uint32) {
	ms, ok := WrapMessage(uuid, message)
	p := NewPackage()
	p.SetACK(ack)
	p.SetRequestCode(requestCode)

	//log.Println("OK: ", ok)
	//if len(ms) <=  LengthHeadData {
	if ok {
		// 数据较短，数据放在头中发送
		p.SetHeadData(ms)
		cli.DataSend <- &p
		return
	}

	filename, hashVal := SaveToFile([]byte(message))
	p.SetExtendedDataFlag(1)
	p.SetExternalDataCheckSum(hashVal)
	PrepareSendFile(&p, cli, def.TempDir, filename)
}

func SaveToFile(data []byte) (filename string, crc32 uint32) {
	filename = RandomFilename()
	err := ioutil.WriteFile(def.TempDir+filename, data, 0644)
	if err != nil {
		return "", 0
	}
	return filename, CRC32(data)
}

// 自适应发送数据，数据较长时自动转为文件发送
func PrepareSendFile(p *Package, cli *Connection, path, filename string) {
	p.SetHeadData([]byte(filename))
	p.SetExtendedDataFlag(1)

	// 计算文件校验值
	fs, err := os.Open(path + filename)
	if err != nil {
		log.Println("os.Open err =", err)
		return
	}
	defer func() {_=fs.Close()}()
	buf := make([]byte, 1024*10)
	n := 0
	hashVal := uint32(0)

	for {
		n, err = fs.Read(buf)
		if err != nil || n == 0 {
			break
		}
		hashVal = crc32.Update(hashVal, crc32.IEEETable, buf[:n])
	}
	p.SetExternalDataCheckSum(hashVal)

	// 将信息打包以filename为key存储在 FileMap 中
	info := &FileMapInfo{
		Path: path,
		IP:       GetRemoteIpUseConn(cli.Connection),
		From:     0,
		To:       []FileMapInfoTo{{GetRemoteIpUseConn(cli.Connection), cli.UUID}},
		CheckSum: p.GetExternalDataCheckSum(),
		Mutex: sync.Mutex{},
	}

	FileMapMutex.Lock()
	FileMap[filename] = info
	FileMapMutex.Unlock()

	cli.DataSend <- p
}

func CalculateFileHashValue(path string) (filename, filePath string, hash uint32) {
	val := uint32(0)
	fs, err := os.Open(path)
	if err != nil {
		log.Println("os.Open err =", err)
		return "", "", 0
	}
	defer func() {_=fs.Close()}()
	info, err := fs.Stat()
	if err != nil {
		return "", "", 0
	}
	filename = info.Name()
	filePath = path[:len(path)-len(filename)]
	buf := make([]byte, 1024*10)
	var n int
	for {
		n, err = fs.Read(buf)
		if err != nil {
			break
		}
		val = crc32.Update(val, crc32.IEEETable, buf[:n])
	}
	return filename, filePath, val
}

func PrintPackage(p *Package, per bool, rec bool) {
	if !def.DisplayPackageInfo && !per {
		return
	}
	if rec {
		fmt.Println("*************REC*************")
	} else {
		fmt.Println("*************SED*************")
	}
	fmt.Println("-----------Package-----------")
	fmt.Println("TIME:", p.GetTime(), time.Unix(0, int64(p.GetTime())).Format("2006-01-02 15:04:05"))
	fmt.Println("CODE:", p.GetRequestCode())
	fmt.Println("SEQ :", p.GetSEQ())
	fmt.Println("ACK :", p.GetACK())
	fmt.Println("FLAG:", p.GetExtendedDataFlag())
	fmt.Println("DATA:", p.Data())
	fmt.Println("-----------------------------")
}

func InfoPrintf(format string, v ...interface{}) {
	if !def.DisplayInfo {
		return
	}
	log.Printf(format, v...)
}

func InfoPrintln(v ...interface{}) {
	log.Println(v...)
}
