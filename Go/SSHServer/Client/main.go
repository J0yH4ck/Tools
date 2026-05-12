package main

// 使用以下命令进行编译，并运行，输入 ip:port 便可以连接到服务器
// go build main.go

// 需要在这个文件里配置用户名和密码

// 运行后出现乱码，可以输入 chcp 65001 便可以解决乱码

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	// 服务端地址和端口
	server := "127.0.0.1:11122"
	fmt.Print("[*] 请输入服务器地址（格式：ip:端口）: ")
	_, err := fmt.Scanln(&server)

	// SSH 连接配置
	config := &ssh.ClientConfig{
		User: "", // 用户名，服务端无需认证时可任意指定
		Auth: []ssh.AuthMethod{
			ssh.Password(""), // 密码
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 连接到服务端
	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatalf("[-] Failed to dial: %s", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("[-] Failed to create session: %s", err)
	}
	defer session.Close()

	// 配置会话的输入、输出和错误流到本地终端
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	if err = session.RequestPty("xterm-256color", 32, 160, modes); err != nil {
		log.Fatalf("request pty error: %s", err.Error())
	}

	// 启动交互式 shell
	fmt.Println("[*] Starting interactive shell")
	err = session.Shell()
	if err != nil {
		log.Fatalf("[-] Failed to start shell: %s", err)
	}

	// 等待 shell 结束
	err = session.Wait()
	if err != nil {
		log.Fatalf("[-] Shell exited with error: %s", err)
	}

	fmt.Println("[+] Interactive shell session ended")
}
