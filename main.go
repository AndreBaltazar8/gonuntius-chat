package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AndreBaltazar8/gonuntius"

	"github.com/AndreBaltazar8/autorpc"
)

type remoteAPI struct {
	Message func(string, func(error))
}

type local struct {
}

func (local *local) Message(str string) {
	fmt.Println("other:", str)
}

func initChat(conn gonuntius.RemoteConnection, first bool) {
	fmt.Println("connected to", string(conn.GetRemoteID()))

	service := autorpc.NewServiceBuilder(&local{}).UseRemote(remoteAPI{}).Build()
	var remote *remoteAPI

	go func() {
		err := service.HandleConnection(conn, func(connection autorpc.Connection) {
			remoteVal, err := connection.GetValue(remoteAPI{})
			if err != nil {
				panic("could not get remote api for client")
			}

			if remoteVal, ok := remoteVal.(*remoteAPI); ok {
				remote = remoteVal
			} else {
				panic("unknown type returned for remote api")
			}
		})

		if err != nil {
			fmt.Println("reading terminated:", err)
			return
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		var msg string
		fmt.Print("msg: ")
		msg, _ = reader.ReadString('\n')
		remote.Message(msg, func(error) {
			fmt.Println("delivered")
		})
	}
}

func main() {
	conn, err := gonuntius.NewConnection()
	if err != nil {
		log.Fatal(err)
	}

	conn.OnReady(func() {
		fmt.Println("conn is ready")
		/*conn.Register([]byte("myApp"), []byte("AndreBaltazar"), []byte("myRegKey"), func(secretKey []byte, err error) {
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("registration success", secretKey)
		})*/
		var id string
		fmt.Print("id: ")
		fmt.Scanf("%s\n", &id)
		conn.Authenticate([]byte{1}, []byte(id), []byte{1}, func(error) {
			fmt.Println("authenticated")

			for {
				var cmd string
				fmt.Print("cmd: ")
				fmt.Scanf("%s\n", &cmd)
				if strings.HasPrefix(cmd, "c") {
					idConn := strings.TrimPrefix(cmd, "c")
					fmt.Println("connecting to", idConn)
					conn.ConnectTo([]byte(idConn), func(remoteConn gonuntius.RemoteConnection, err error) {
						if err != nil {
							fmt.Println("failed to connect:", err)
							return
						}

						initChat(remoteConn, true)
					})
					break
				} else if cmd == "wait" {
					break
				}
			}
		})
	})

	conn.OnIncomingConnection(func(inc gonuntius.IncomingConnection) {
		fmt.Println("connection from", string(inc.GetRemoteID()))
		inc.Accept(func(conn gonuntius.RemoteConnection, err error) {
			if err != nil {
				fmt.Println("failed to connect:", err)
				return
			}

			initChat(conn, false)
		})
	})

	for {
		time.Sleep(1 * time.Second)
	}
}
