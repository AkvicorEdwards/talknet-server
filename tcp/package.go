package tcp

const (
	// 数据最大长度
	LengthHeadPackage     = 210
	// 偏移
	OffsetRequestCode = 0
	LengthRequestCode = 2
	OffsetSEQ = 2
	LengthSEQ = 4
	OffsetACK = 6
	LengthACK = 4
	OffsetTime = 10
	LengthTime = 8
	OffsetExtendedDataFlag = 18
	LengthExtendedDataFlag = 1
	OffsetHeadDataLength = 19
	LengthHeadDataLength = 1
	OffsetHeadData = 20
	LengthHeadData = 182
	OffsetExtendedDataHash = 202
	LengthExtendedDataHash = 4
	OffsetHash = 206
	LengthHash = 4
)

type Package struct {
	data [LengthHeadPackage]byte
}

// 将byte数组转化为Package
func ConvertToPackage(data []byte) (p Package) {
	if len(data) < LengthHeadPackage {
		copy(p.data[:], data[:])
	}
	copy(p.data[:], data[:LengthHeadPackage])
	return
}

func NewPackage() Package {
	return Package{data: [LengthHeadPackage]byte{}}
}

// 获取data的内容
func (p *Package) Data() []byte {
	return p.data[:]
}

// 设置请求代码
func (p *Package) SetRequestCode(request uint16) {
	data := UInt16ToBytes(request)
	for k, v := range data {
		p.data[OffsetRequestCode+k] = v
	}
}

// 获取请求代码
func (p *Package) GetRequestCode() uint16 {
	return BytesToUInt16(p.data[:OffsetRequestCode+LengthRequestCode])
}

// 清除请求代码
func (p *Package) ClearRequestCode() {
	for i := OffsetRequestCode; i < OffsetRequestCode+LengthRequestCode; i++ {
		p.data[i] = 0
	}
}

// 设置序列号
func (p *Package) SetSEQ(seq uint32) {
	data := UInt32ToBytes(seq)
	for k, v := range data {
		p.data[OffsetSEQ+k] = v
	}
}

// 获取序列号
func (p *Package) GetSEQ() uint32 {
	return BytesToUInt32(p.data[OffsetSEQ:OffsetSEQ+LengthSEQ])
}

// 清除序列号
func (p *Package) ClearSEQ() {
	for i := OffsetSEQ; i < OffsetSEQ+LengthSEQ; i++ {
		p.data[i] = 0
	}
}

// 设置响应序列号
func (p *Package) SetACK(ack uint32) {
	data := UInt32ToBytes(ack)
	for k, v := range data {
		p.data[OffsetACK+k] = v
	}
}

// 获取响应序列号
func (p *Package) GetACK() uint32 {
	return BytesToUInt32(p.data[OffsetACK:OffsetACK+LengthACK])
}

// 清除响应序列号
func (p *Package) ClearACK() {
	for i := OffsetACK; i < OffsetACK+LengthACK; i++ {
		p.data[i] = 0
	}
}

// 设置时间戳
func (p *Package) SetTime(t uint64) {
	data := UInt64ToBytes(t)
	for k, v := range data {
		p.data[OffsetTime+k] = v
	}
}

// 获取时间戳
func (p *Package) GetTime() uint64 {
	return BytesToUInt64(p.data[OffsetTime:OffsetTime+LengthTime])
}

// 清除时间戳
func (p *Package) ClearTime() {
	for i := OffsetTime; i < OffsetTime+LengthTime; i++ {
		p.data[i] = 0
	}
}

// 设置额外数据标签
// 代表有额外数据需要新建tcp连接来接收
func (p *Package) SetExtendedDataFlag(flag byte) {
	p.data[OffsetExtendedDataFlag] = flag
}

// 获取额外数据标签
func (p *Package) GetExtendedDataFlag() byte {
	return p.data[OffsetExtendedDataFlag]
}

// 清除额外数据标签
func (p *Package) ClearExtendedDataFlag() {
	for i := OffsetExtendedDataFlag; i < OffsetExtendedDataFlag+LengthExtendedDataFlag; i++ {
		p.data[i] = 0
	}
}

// 设置头内数据
func (p *Package) SetHeadData(data []byte) {
	if len(data) > LengthHeadData {
		data = data[:LengthHeadData]
	}
	p.data[OffsetHeadDataLength] = byte(len(data))
	for k, v := range data {
		p.data[OffsetHeadData+k] = v
	}
}

// 获取头内数据
func (p *Package) GetHeadData() []byte {
	return p.data[OffsetHeadData:OffsetHeadData+p.data[OffsetHeadDataLength]]
}

// 清除头内数据
func (p *Package) ClearHeadData() {
	for i := OffsetHeadDataLength; i < OffsetHeadDataLength+LengthHeadDataLength; i++ {
		p.data[i] = 0
	}
	for i := OffsetHeadData; i < OffsetHeadData+LengthHeadData; i++ {
		p.data[i] = 0
	}
}

// 设置额外数据的哈希值
func (p *Package) SetExternalDataCheckSum(sum uint32) {
	data := UInt32ToBytes(sum)
	for k, v := range data {
		p.data[OffsetExtendedDataHash+k] = v
	}
}

// 获取额外数据的哈希值
func (p *Package) GetExternalDataCheckSum() uint32 {
	return BytesToUInt32(p.data[OffsetExtendedDataHash:OffsetExtendedDataHash+LengthExtendedDataHash])
}

// 清除额外数据的哈希值
func (p *Package) ClearExternalDataCheckSum() {
	for i := OffsetExtendedDataHash; i < OffsetExtendedDataHash+LengthExtendedDataHash; i++ {
		p.data[i] = 0
	}
}

// 检查额外数据的哈希值
func (p *Package) CheckExternalDataCheckSum(sum uint32) bool {
	return sum == p.GetExternalDataCheckSum()
}

// 设置package的哈希值
func (p *Package) SetHeadCheckSum() {
	data := UInt32ToBytes(CRC32(p.data[:OffsetHash]))
	for k, v := range data {
		p.data[OffsetHash+k] = v
	}
}

// 获取package的哈希值
func (p *Package) GetHeadCheckSum() uint32 {
	return BytesToUInt32(p.data[OffsetHash:OffsetHash+LengthHash])
}

// 清除package的哈希值
func (p *Package) ClearHeadCheckSum() {
	for i := OffsetHash; i < OffsetHash+LengthHash; i++ {
		p.data[i] = 0
	}
}

// 检查package的哈希值
func (p *Package) CheckHeadCheckSum() bool {
	//log.Println(CRC32(p.data[:OffsetHash]))
	//log.Println(p.GetHeadCheckSum())
	return CRC32(p.data[:OffsetHash]) == p.GetHeadCheckSum()
}

// 重置package
func (p *Package) Clear() {
	p.ClearRequestCode()
	p.ClearSEQ()
	p.ClearACK()
	p.ClearTime()
	p.ClearExtendedDataFlag()
	p.ClearHeadData()
	p.ClearHeadCheckSum()
	p.ClearExternalDataCheckSum()
}

// 清除SEQ以外的数据
func (p *Package) ClearExceptSeq() {
	p.ClearRequestCode()
	p.ClearACK()
	p.ClearTime()
	p.ClearExtendedDataFlag()
	p.ClearHeadData()
	p.ClearHeadCheckSum()
	p.ClearExternalDataCheckSum()
}
