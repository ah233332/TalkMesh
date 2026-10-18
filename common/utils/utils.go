package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

// readPkg 从连接中读取一个完整的数据包。
// 它先读取 4 字节大端长度头，再按该长度读取消息体，
// 最后将 JSON 格式的消息体反序列化为 message.Message。
func ReadPkg(conn net.Conn) (mes message.Message, err error) {
	buf := make([]byte, 8096)
	fmt.Println("读取客户端发送的数据...")
	_, err = conn.Read(buf[:4])
	if err != nil {
		fmt.Println("conn read err=", err)
		return
	}

	//把buf[:4]转成一个uint32类型
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(buf[:4])

	//根据pkgLen读取消息内容
	n, err := conn.Read(buf[:pkgLen]) //从conn读取pkgLen个字节到buf里面
	if n != int(pkgLen) || err != nil {
		fmt.Println("conn read err=", err)
		return
	}

	//buf  反序列化成 Message类型
	err = json.Unmarshal(buf[:pkgLen], &mes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	return

}

// 通用的打包发送函数，先发送数据长度，再发送数据本体
func WritePkg(conn net.Conn, data []byte) (err error) {
	//1.先发送长度给对方
	var dataLen uint32
	dataLen = uint32(len(data))

	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:4], dataLen)
	n, err := conn.Write(buf[:4])
	if err != nil || n != 4 {
		fmt.Println("发送信息长度失败=", err)
		return
	}
	//2.再发送本体数据给对方
	n, err = conn.Write(data)
	if n != int(dataLen) || err != nil {
		fmt.Println("发送信息本体失败=", err)
		return
	}
	return
}
