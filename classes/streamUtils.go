//通讯协议处理，主要处理封包和解包的过程
package streamtcp

import (
	"bytes"
	"encoding/binary"
	//"fmt"
	"hash/adler32"
	//"runtime"
	"fmt"
	"strings"
)

const (
	ConstHeader                  = "SeanCommu"
	ConstHeaderLength      int32 = 9
	ConstSaveDataLength    int32 = 4
	ConstEndCheckECCLength int32 = 4

	TempToken = "e675d1b7e23c8a9ca3a3479d6cf29232"
)

const (
	ConstPackNoHeader  = 0
	ConstPackHasHeader = 1
)

var (
	num = make(chan int, 100)
)

func init() {
	go func() {
		for i := 0; ; i++ {
			num <- i
		}
	}()
}

const (
	NAME_PREFIX = "User "
)

func getUniqName() string {
	return fmt.Sprintf("%s%d", NAME_PREFIX, GetUniq())
}

func GetUniq() int {
	return <-num
}

func CheckECC(data []byte) uint32 {
	if len(data) > 200 {
		tmpdata := append(data[:100], data[len(data)-100:]...)
		return adler32.Checksum(tmpdata)
	} else {
		return adler32.Checksum(data)
	}
}

//封包
func Packet(message []byte) []byte {
	data := append(append([]byte(ConstHeader), IntToBytes(int32(len(message)))...), message...)
	return append(data, UInt32ToBytes(CheckECC(data))...)
}

//解包
func Unpack(buffer []byte, readerChannel chan []byte) (leftbuffer []byte) {
	length := int32(len(buffer))

	var i int32

	for i = 0; i < length; i = i + 1 {
		//如果当前缓冲区中的数据小于包头长度 ,不需要继续向下寻找包头了，寻找就越界了
		if length < i+ConstHeaderLength+ConstSaveDataLength {
			break
		}
		//if bytes.Contains(buffer[i:i+ConstHeaderLength], []byte(ConstHeader)) {
		if strings.EqualFold(string(buffer[i:i+ConstHeaderLength]), ConstHeader) { //string(buffer[i:i+ConstHeaderLength]) == []byte(ConstHeader) {
			//token = string(buffer[i+ConstHeaderLength+ConstTokenSizeLength : i+ConstHeaderLength+ConstTokenSizeLength+ConstTokenLength])
			//fmt.Println(TokenStr)
			//TokenStr
			messageLength := BytesToInt(buffer[i+ConstHeaderLength : i+ConstHeaderLength+ConstSaveDataLength])

			if length < i+ConstHeaderLength+ConstSaveDataLength+messageLength+ConstEndCheckECCLength {
				break
			}
			data := buffer[i+ConstHeaderLength+ConstSaveDataLength : i+ConstHeaderLength+ConstSaveDataLength+messageLength]
			ECCSourceData := buffer[i : i+ConstHeaderLength+ConstSaveDataLength+messageLength]
			ECCSum := BytesTouInt32(buffer[i+ConstHeaderLength+ConstSaveDataLength+messageLength : i+ConstHeaderLength+ConstSaveDataLength+messageLength+ConstEndCheckECCLength])
			ECCCalcSum := CheckECC(ECCSourceData)
			if ECCSum != ECCCalcSum {
				i += i + ConstHeaderLength - 1
				continue
			}
			readerChannel <- data

			i += ConstHeaderLength + ConstSaveDataLength + messageLength + ConstEndCheckECCLength - 1

		}
	}

	if i == length {
		//return buffer, ConstPackNoHeader
		return make([]byte, 0)
	}
	//runtime.GC()
	return buffer[i:]
}

//整形转换成字节
func IntToBytes(n int32) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

//整形转换成字节
func UInt32ToBytes(n uint32) []byte {
	x := uint32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesTouInt32(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)

	var x uint32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)

	return uint32(x)
}

//字节转换成整形
func BytesToInt(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)

	return int32(x)
}
