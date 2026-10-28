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

// 通过 t.Conn.Read 先读取 4 字节大端长度头并转换为 pkgLen，再按该长度把 JSON 消息体读入 t.Buf。
// 读取完成后使用 json.Unmarshal 将数据写入 mes 并返回；长度不符、连接读取失败或解析失败时返回 err。
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

// 通过 len(data) 计算 dataLen，使用大端序把长度写入 t.Buf[:4] 并先发送长度头。
// 随后继续发送 data 本体，并检查每次写入的字节数与 err，确保接收端能够按相同边界拆分消息。
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
