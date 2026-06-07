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
			} else {
				fmt.Printf("  ➜ Reenviado para %s\n", endereco)
			}
		}
	}
}

func iniciarServidor(endereco string) {
	fmt.Println("🎮 Iniciando o túnel multiplayer (SERVIDOR)...")

	addr, err := net.ResolveUDPAddr("udp", endereco)
	if err != nil {
		fmt.Println("Erro ao resolver endereço:", err)
		os.Exit(1)
	}

	conexao, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Erro ao abrir a porta UDP:", err)
		os.Exit(1)
	}
	defer conexao.Close()

	fmt.Printf("✅ Ouvindo em %s... Pronto para receber dados.\n", endereco)
	fmt.Println("⏳ Aguardando conexão de jogadores...\n")

	buffer := make([]byte, 4096)

	for {
		n, remoteAddr, err := conexao.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Erro ao capturar os pacotes de dados:", err)
			os.Exit(1)
		}

		mutex.Lock()
		clientesConectados[remoteAddr.String()] = remoteAddr
		mutex.Unlock()

		mensagem := string(buffer[:n])
		fmt.Printf("📍 Recebido de %s: %s\n", remoteAddr.String(), mensagem)
		fmt.Printf("👥 Clientes ativos: %d\n\n", len(clientesConectados))
		
		retransmitirParaOutros(conexao, buffer[:n], remoteAddr)
	}
}

func iniciarCliente(enderecoServidor string) {
	fmt.Println("🎮 Iniciando o túnel multiplayer (CLIENTE)...")
	fmt.Printf("🔗 Conectando a %s...\n\n", enderecoServidor)

	
	serverAddr, err := net.ResolveUDPAddr("udp", enderecoServidor)
	if err != nil {
		fmt.Println("❌ Erro ao resolver servidor:", err)
		os.Exit(1)
	}

	
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		fmt.Println("❌ Erro ao resolver endereço local:", err)
		os.Exit(1)
	}

	conexao, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		fmt.Println("❌ Erro ao conectar ao servidor:", err)
		os.Exit(1)
	}
	defer conexao.Close()

	fmt.Println("✅ Conectado ao servidor!")
	fmt.Println("📤 Seu Minecraft agora está tunnelizado!\n")

	
	go receberDadosCliente(conexao)

	
	enviarDadosCliente(conexao)
}

func receberDadosCliente(conexao *net.UDPConn) {
	buffer := make([]byte, 4096)

	for {
		n, _, err := conexao.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("❌ Erro ao receber dados:", err)
			return
		}

		
		fmt.Printf("📥 Recebido %d bytes de outro jogador\n", n)
	}
}

func enviarDadosCliente(conexao *net.UDPConn) {
	buffer := make([]byte, 4096)

	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			fmt.Println("❌ Erro ao ler entrada:", err)
			return
		}

		_, err = conexao.Write(buffer[:n])
		if err != nil {
			fmt.Println("❌ Erro ao enviar:", err)
			return
		}

		fmt.Printf("📤 Enviado %d bytes ao servidor\n", n)
	}
}

func main() {
	modo := flag.String("modo", "host", "host ou join")
	endereco := flag.String("endereco", "0.0.0.0:9999", "endereço para escutar ou conectar")
	flag.Parse()

	if *modo == "host" {
		iniciarServidor(*endereco)
	} else if *modo == "join" {
		iniciarCliente(*endereco)
	} else {
		fmt.Println("❌ Modo inválido. Use -modo=host ou -modo=join")
		os.Exit(1)
	}
}
