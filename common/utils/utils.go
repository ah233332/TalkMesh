package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

type Transfer struct {
	Conn net.Conn
	Buf  [8096]byte
}

// ReadPkg 从 t.Conn 读取一个完整的数据包。
// 先用 io.ReadFull 读满 4 字节大端长度头并转换为 pkgLen，校验 pkgLen 不超过缓冲区大小（当前 8096 字节），
// 再用 io.ReadFull 读满包体，最后把 JSON 反序列化为 message.Message。
// 连接读取失败、长度非法或反序列化失败时返回 error，调用方应关闭该连接。
func (t *Transfer) ReadPkg() (mes message.Message, err error) {
	_, err = io.ReadFull(t.Conn, t.Buf[:4])
	if err != nil {
		err = fmt.Errorf("t.Conn ioreadfull err=%w", err)
		return
	}

	//把buf[:4]转成一个uint32类型
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(t.Buf[:4])
	//binary 包实现了数字与字节序列之间的简单转换
	//判断pkglen是否合法
	if int(pkgLen) == 0 || int(pkgLen) > len(t.Buf) {
		err = fmt.Errorf("pkgLen长度错误:%d", int(pkgLen))
		return
	}

	//根据pkgLen读取消息内容
	//io.ReadFull 要么读满成功，要么返回错误
	_, err = io.ReadFull(t.Conn, t.Buf[:pkgLen]) //从conn读取pkgLen个字节到buf里面
	if err != nil {
		err = fmt.Errorf("t.Conn ioreadfull err=%w", err)
		return
	}

	//t.buf  反序列化成 Message类型
	err = json.Unmarshal(t.Buf[:pkgLen], &mes)
	if err != nil {
		err = fmt.Errorf("unmarshal err=%w", err)
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

	//再次验证一下包长
	if int(dataLen) == 0 || int(dataLen) > len(t.Buf) {
		err = fmt.Errorf("dataLen长度错误:%d", int(dataLen))
		return
	}

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
