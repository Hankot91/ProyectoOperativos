package main



import (

	"bufio"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"time"
)



var socketServ net.Conn
var salirServ bool
var tiempoSegundos int

// Estructura para almacenar la configuración del servidor

type ServerConfig struct {

	Port    int
	IP      string
	Users   []string
	DBUsers string
}



// Estructura para almacenar las credenciales de los usuarios

type UserCredentials struct {

	Username string
	Password string
}



func main() {

	fmt.Println("################################")
	fmt.Println("# Servidor Shell/Comandos Unix #")
	fmt.Println("################################")



	// Cargar configuración del archivo
	config := cargarConfig("serverCommands.config")

	// Cargar credenciales de usuarios desde el archivo DBUsers
	credenciales := cargarCredenciales(config.DBUsers)



	//*** ESTABLECIENDO CONEXIÓN ***
	tcpAddress, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", config.Port))
	socketServer, _ := net.ListenTCP("tcp", tcpAddress)
	fmt.Println("\nServ# Esperando conexión con el cliente ")

	socketServ, _ = socketServer.Accept()
	fmt.Println("Serv# Cliente conectando!", socketServ.RemoteAddr())
	// ------ fin establecer conexión ---- //



	for {

		login := login(socketServ, credenciales)
		if login == "correcto" {
			eML := bufio.NewWriter(socketServ)
			mReporte := "correcto"
			eML.WriteString(mReporte + "\n")
			eML.Flush()
			break
		}

		eML := bufio.NewWriter(socketServ)
		mReporte := "Las credenciales son incorrectas"
		eML.WriteString(mReporte + "\n")
		eML.Flush()
	}

	tiempoSegundosStr, _ := bufio.NewReader(socketServ).ReadString('\n')
	tiempoSegundosStr = strings.TrimSpace(tiempoSegundosStr)
	tiempoSegundos, _ := strconv.Atoi(tiempoSegundosStr)

	go recComando(&socketServ)
	go envReporte(&socketServ, tiempoSegundos)



	for {
		if salirServ {
			break
		}
	}

	fmt.Println("=======  CERRANDO SERVIDOR  =========")
	socketServ.Close()
	fmt.Println("=====================================")
	fmt.Println("|| Gracias por usar Shell/Comandos ||")
	fmt.Println("=====================================")
}



func cargarConfig(filename string) *ServerConfig {
	config := &ServerConfig{}
	fileContent, _ := ioutil.ReadFile(filename)
	lines := strings.Split(string(fileContent), "\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Port":
			fmt.Sscanf(value, "%d", &config.Port)
		case "IP":
			config.IP = value
		case "Users":
			config.Users = strings.Split(value, ",")
		case "DBUsers":
			config.DBUsers = value
		}
	}

	return config
}



func cargarCredenciales(filename string) []UserCredentials {
	var credenciales []UserCredentials

	fileContent, _ := ioutil.ReadFile(filename)
	lines := strings.Split(string(fileContent), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			continue
		}
		cred := UserCredentials{
			Username: parts[0],
			Password: parts[3],
		}
		credenciales = append(credenciales, cred)
	}

	return credenciales
}



func login(socketServ net.Conn, credenciales []UserCredentials) string {
	var usuariVald, passwVald bool


	reader := bufio.NewReader(socketServ)
	usuario, _ := reader.ReadString('\n')
	usuario = strings.TrimSpace(usuario)
	passwd, _ := reader.ReadString('\n')
	passwd = strings.TrimSpace(passwd)
	datoInArray := sha256.Sum256([]byte(passwd))
	passUser := fmt.Sprintf("%x", datoInArray)

	for _, cred := range credenciales {
		if usuario == cred.Username {
			fmt.Println("USUARIO ENCONTRADO")
			usuariVald = true
			if passUser == cred.Password {
				fmt.Println("PASS ENCONTRADO")
				passwVald = true
			}
		}
	}


	if usuariVald && passwVald {
		return "correcto"
	}

	return "incorrecto"
}



func recComando(socketServ *net.Conn) {
	for {
		mR, _ := bufio.NewReader(*socketServ).ReadString('\n')
		mR = strings.TrimSpace(mR)

		fmt.Println("Serv# Comando recibido: ", mR)
		if mR == "bye" {
			salirServ = true
			break
		} else if mR != "" && mR != "bye" {
			fmt.Println("Serv# Ejecutando comando: ", mR)
			datoIn := strings.Fields(mR)
			shell := exec.Command(datoIn[0], datoIn[1:]...)
			datoOut, _ := shell.Output()
			sDatoOut := string(datoOut)
			env := bufio.NewWriter(*socketServ)
			env.WriteString(sDatoOut + "\n")
			env.Flush()
			fmt.Println("Serv# Respuesta de ejecución enviada!")
		}
		
	}
}



func obtenerPorcentajeCPU(topInfo string) string {

	lines := strings.Split(topInfo, "\n")
	if len(lines) >= 3 {
		fields := strings.Fields(lines[2])
		if len(fields) >= 10 {
			return "CPU: " + fields[1] + "%"
		}
	}

	return "No se pudo obtener el porcentaje de uso del procesador."
}



func envReporte(socketServ *net.Conn, tiempoSegundos int) {

	for {
		time.Sleep(time.Duration(tiempoSegundos) * time.Second)
		eR := bufio.NewWriter(*socketServ)

		// Obtener información sobre la memoria
		cmdFree := exec.Command("free", "-h")
		cmdAwk := exec.Command("awk", "NR==2 {print $3}")
		outputFree, _ := cmdFree.Output()
		cmdAwk.Stdin = strings.NewReader(string(outputFree))
		outputAwk, _ := cmdAwk.Output()
		mReporte := strings.TrimSpace(string(outputAwk))

		// Obtener información sobre el porcentaje de uso del procesador
		cmdTop := exec.Command("top", "-bn1")
		outputTop, _ := cmdTop.Output()
		cpuUsage := obtenerPorcentajeCPU(string(outputTop))

		// Obtener información sobre el disco
		cmdDf := exec.Command("df", "-h")
		cmdAwk1 := exec.Command("awk", "NR==2 {print $2}")
		outputDf, _ := cmdDf.Output()
		cmdAwk1.Stdin = strings.NewReader(string(outputDf))
		outputAwk1, _ := cmdAwk1.Output()
		mReporte1 := strings.TrimSpace(string(outputAwk1))

		// Escribir la información en el cliente
		eR.WriteString(cpuUsage + ", memoria: " +  string(mReporte) + ", Disco: " + string(mReporte1) + "\n" )
		eR.Flush()
		fmt.Println("Serv# Reporte enviado al cliente")
	}
}

