package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Node struct {
	Connections map[string]bool
	Address     Address
}

type Address struct {
	IPv4 string
	Port string
}

//информация - от кого и кому отправлено сообщение
type Package struct {
	To   string
	From string
	Data string
}

//функция инициализации, где пользователь вводит свой адрес
// ./main :8080
func init() {
	//если аргумента не 2 (айпи адрес и порт) то выкинуть
	if len(os.Args) != 2 {
		panic("arguments less than two")
	}
}

func main() {
	NewNode(os.Args[1]).Run(handleServer, handleClient)
}

//ipv4:port
func NewNode(address string) *Node {
	splited := strings.Split(address, ":")
	if len(splited) != 2 {
		return nil
	}
	return &Node{
		Connections: make(map[string]bool),
		Address: Address{
			IPv4: splited[0],
			Port: ":" + splited[1],
		},
	}
}

func (node *Node) Run(handleServer func(*Node), handleClient func(*Node)) {
	go handleServer(node)
	handleClient(node)
}

//сервер на принятие данных с определенного порта
func handleServer(node *Node) {
	listen, err := net.Listen("tcp", "0.0.0.0"+node.Address.Port)
	if err != nil {
		panic("listen error")
	}
	defer listen.Close()

	for {
		//приятие соединения
		conn, err := listen.Accept()
		if err != nil {
			break
		}
		go handleConnection(node, conn)
	}
}

func handleConnection(node *Node, conn net.Conn) {
	defer conn.Close()
	var (
		buffer  = make([]byte, 512)
		message string
		pack    Package
	)
	for {
		length, err := conn.Read(buffer)
		if err != nil {
			break
		}
		message += string(buffer[:length])
	}
	//расшифровка JSON файла
	err := json.Unmarshal([]byte(message), &pack)
	if err != nil {
		return
	}
	//проанализировать от кого отправлен пакет
	node.ConnectTo([]string{pack.From})
	//смотрим что отправил пользователь
	fmt.Println(pack.Data)
}

func handleClient(node *Node) {
	for {
		message := InputString()
		splited := strings.Split(message, " ")
		switch splited[0] {
		case "/exit":
			os.Exit(0)
		case "/connect":
			node.ConnectTo(splited[1:])
		case "/network":
			node.PrintNetWork()
		default:
			node.SendMessageToAll(message)
		}
	}
}

func (node *Node) PrintNetWork() {
	//перебор адресов в соединениях
	for addr := range node.Connections {
		fmt.Println("|", addr)
	}
}

func (node *Node) ConnectTo(addresses []string) {
	//подлючение к клиентам через перебор адресов
	for _, addr := range addresses {
		node.Connections[addr] = true
	}
}

func (node *Node) SendMessageToAll(message string) {
	var new_pack = Package{
		From: node.Address.IPv4 + node.Address.Port,
		Data: message,
	}
	for addr := range node.Connections {
		new_pack.To = addr
		node.Send(&new_pack)
	}
}

func (node *Node) Send(pack *Package) {
	conn, err := net.Dial("tcp", pack.To)
	if err != nil { // пытаемся удалить адресс из соединения, если он существует
		delete(node.Connections, pack.To)
		return
	}
	defer conn.Close()                  //закрытие соединения
	json_pack, _ := json.Marshal(*pack) //перевод в JSON формат
	conn.Write(json_pack)
}

func InputString() string {
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", -1)
}