package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
)

var clientesConectados = make(map[string]*net.UDPAddr)
var mutex = sync.RWMutex{}

func retransmitirParaOutros(conexao *net.UDPConn, dados []byte, origem *net.UDPAddr) {
	mutex.RLock()
	defer mutex.RUnlock()

	for endereco, cliente := range clientesConectados {
		if endereco != origem.String() {
			_, err := conexao.WriteToUDP(dados, cliente)
			if err != nil {
				fmt.Println("Erro ao reenviar para", endereco, ":", err)
			}
		}
	}
}

func iniciarServidor(endereco string) {
	fmt.Println("Iniciando servidor do tunel...")

	addr, err := net.ResolveUDPAddr("udp", endereco)
	if err != nil {
		fmt.Println("Erro ao resolver endereco:", err)
		os.Exit(1)
	}

	conexao, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Erro ao abrir a porta UDP:", err)
		os.Exit(1)
	}
	defer conexao.Close()

	fmt.Printf("Servidor ouvindo em %s\n", endereco)
	fmt.Println("Aguardando conexoes...\n")

	buffer := make([]byte, 4096)

	for {
		n, remoteAddr, err := conexao.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Erro ao receber dados:", err)
			continue
		}

		mutex.Lock()
		clientesConectados[remoteAddr.String()] = remoteAddr
		mutex.Unlock()

		fmt.Printf("Cliente conectado: %s (Total: %d)\n", remoteAddr.String(), len(clientesConectados))
		
		retransmitirParaOutros(conexao, buffer[:n], remoteAddr)
	}
}

func iniciarCliente(enderecoServidor string) {
	fmt.Println("Conectando ao servidor do tunel...")
	fmt.Printf("Endereco: %s\n\n", enderecoServidor)

	serverAddr, err := net.ResolveUDPAddr("udp", enderecoServidor)
	if err != nil {
		fmt.Println("Erro ao resolver servidor:", err)
		os.Exit(1)
	}

	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		fmt.Println("Erro ao resolver endereco local:", err)
		os.Exit(1)
	}

	conexao, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		os.Exit(1)
	}
	defer conexao.Close()

	fmt.Println("Conectado ao servidor!")
	fmt.Println("Tunel ativo. Pode conectar no Minecraft.\n")

	go receberDadosCliente(conexao)
	enviarDadosCliente(conexao)
}

func receberDadosCliente(conexao *net.UDPConn) {
	buffer := make([]byte, 4096)

	for {
		_, _, err := conexao.ReadFromUDP(buffer)
		if err != nil {
			return
		}
	}
}

func enviarDadosCliente(conexao *net.UDPConn) {
	buffer := make([]byte, 4096)

	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			return
		}

		_, err = conexao.Write(buffer[:n])
		if err != nil {
			return
		}
	}
}

func main() {
	modo := flag.String("modo", "host", "host ou join")
	endereco := flag.String("endereco", "0.0.0.0:9999", "endereco para escutar ou conectar")
	flag.Parse()

	if *modo == "host" {
		iniciarServidor(*endereco)
	} else if *modo == "join" {
		iniciarCliente(*endereco)
	} else {
		fmt.Println("Modo invalido. Use -modo=host ou -modo=join")
		os.Exit(1)
	}
}
