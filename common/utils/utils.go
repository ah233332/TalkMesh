package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

type Transfer struct {
	Conn net.Conn
	Buf  [8096]byte
}

// readPkg 从连接中读取一个完整的数据包。
// 它先读取 4 字节大端长度头，再按该长度读取消息体，
// 最后将 JSON 格式的消息体反序列化为 message.Message。
func (t *Transfer) ReadPkg() (mes message.Message, err error) {
	fmt.Println("读取客户端发送的数据...")
	_, err = t.Conn.Read(t.Buf[:4])
	if err != nil {
		fmt.Println("t.Conn read err=", err)
		return
	}

	//把buf[:4]转成一个uint32类型
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(t.Buf[:4])

	//根据pkgLen读取消息内容
	n, err := t.Conn.Read(t.Buf[:pkgLen]) //从conn读取pkgLen个字节到buf里面
	if n != int(pkgLen) || err != nil {
		fmt.Println("t.Conn read err=", err)
		return
	}

	//t.buf  反序列化成 Message类型
	err = json.Unmarshal(t.Buf[:pkgLen], &mes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	return

}

// 通用的打包发送函数，先发送数据长度，再发送数据本体
func (t *Transfer) WritePkg(data []byte) (err error) {
	//1.先发送长度给对方
	var dataLen uint32
	dataLen = uint32(len(data))

	binary.BigEndian.PutUint32(t.Buf[:4], dataLen)
	n, err := t.Conn.Write(t.Buf[:4])
	if err != nil || n != 4 {
		fmt.Println("发送信息长度失败=", err)
		return
	}
	//2.再发送本体数据给对方
	n, err = t.Conn.Write(data)
	if n != int(dataLen) || err != nil {
		fmt.Println("发送信息本体失败=", err)
		return
	}
	return
}
