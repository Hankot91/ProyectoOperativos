package main


import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

)

var socketCli net.Conn
var salirCli bool


func main() {

	fmt.Println("################################")
	fmt.Println("# Cliente Shell/Comandos Unix #")
	fmt.Println("################################")


	//*** ESTABLECIENDO CONEXIÓN ***
	//var IPserver string = os.Args[1]
	//var portServer string = os.Args[2]
	fmt.Println("Cli# Estableciendo conexión ...")
	tcpAddres, err := net.ResolveTCPAddr("tcp4", "10.1.10.48"+":"+"8000")
	if err != nil {
		fmt.Println("Error al resolver la dirección:", err)
		
time.Sleep(3 * time.Second)
		return
	}


	socketCli, err = net.DialTCP("tcp", nil, tcpAddres)
	if err != nil {
		fmt.Println("Error al conectar al servidor:", err)
		time.Sleep(3 * time.Second)
		return
	}

	fmt.Println("Cli# conectado con [", socketCli.RemoteAddr(), ":", 8000, "]")
	// ------ fin establecer conexión ---- //


	for {
		eR := bufio.NewWriter(socketCli)
		fmt.Println("Ingresa el usuario")
		reader := bufio.NewReader(os.Stdin)
		userReader, _ := reader.ReadString('\n')
		userReader = strings.TrimSuffix(userReader, "\n")
		eR.WriteString(userReader + "\n")
		eR.Flush()



		fmt.Println("Ingresa la contraseña")
		reader2 := bufio.NewReader(os.Stdin)
		password, _ := reader2.ReadString('\n')
		eR.WriteString(password + "\n")
		eR.Flush()
		acppLog, _ := bufio.NewReader(socketCli).ReadString('\n')
		acppLog = strings.TrimSpace(acppLog)

		fmt.Println("Cli# Reporte recibido del servidor:", acppLog)
		if acppLog == "correcto" {
			break
		}

	}
	var tiempoSegundos int
	fmt.Println("Clie# Ingresa el tiempo en segundos:")
	fmt.Scanln(&tiempoSegundos)

	eML := bufio.NewWriter(socketCli)
	eML.WriteString(fmt.Sprintf("%d\n", tiempoSegundos))
	eML.Flush()

	go envComando(&socketCli)
	go recReporte(&socketCli)

	for {
		if salirCli {
			break
		}
	}


	fmt.Println("============= CERRANDO CONEXIÓN ==============")
	socketCli.Close()
	fmt.Println("==============================================")
	fmt.Println("|| Gracias por usar Cliente Shell /Comandos ||")
	fmt.Println("==============================================")

}


func envComando(socketCli *net.Conn) {

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Cli# Digite el comando a enviar:")
		comandoCli, _ := reader.ReadString('\n')
		comandoCli = strings.TrimSpace(comandoCli)
		if comandoCli == "bye" {
			mE := bufio.NewWriter(*socketCli)
			mE.WriteString(comandoCli + "\n")
			mE.Flush()
			salirCli = true

		}
		mE := bufio.NewWriter(*socketCli)
		mE.WriteString(comandoCli + "\n")
		mE.Flush()
		fmt.Println("Cli# comando enviado al servidor:")
	}

}


func recReporte(socketCli *net.Conn) {

	for {
		rec := bufio.NewReader(*socketCli)
		for {
			sResComando, _ := rec.ReadString('\n')
			fmt.Print("Cliente#", sResComando)
			if sResComando == "\n" {
				break
			}
		}

		time.Sleep(4 * time.Second)
		rR, _ := bufio.NewReader(*socketCli).ReadString('\n')
		fmt.Println("Cli# ", rR)
		fmt.Println("Cli# Puedes seguir ingresando comandos:")
	}

}


func config() {
	conf := leerConfig("serverCommands.config")
	println(conf)
}


func leerConfig(fileName string) []string {
	var resp []string
	archivo, _ := ioutil.ReadFile(fileName)
	sArchivo := string(archivo)
	credenciales := strings.Split(sArchivo, ":")
	for i := 0; i < len(credenciales); i++ {
		resp = append(resp, credenciales[i])
	}

	return resp
}

