package main

import (
	"bufio"
	"fmt"
	"github.com/lionslon/go-keepass/internal/client/app"
	"github.com/lionslon/go-keepass/internal/client/config"
	"log"
	"os"
	"strings"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"

	reader *bufio.Reader
)

//go build -ldflags="-X 'main.Version=v1.0.0' -X 'app/build.Time=$(date)'"

func readLine(title string) string {
	fmt.Printf(`Enter %s: `, title)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSuffix(line, "\r\n")
	return line
}

func main() {

	reader = bufio.NewReader(os.Stdin)

	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	sender := app.NewSender(cfg)
	if err := sender.Init(); err != nil {
		log.Fatalf("cannot initialize sender: %s\n", err)
	}

	for {
		cmd := readLine(`command`)

		switch cmd {
		case `register`:
			login := readLine(`login`)
			password := readLine(`password`)

			err := sender.Register(login, password)
			if err != nil {
				fmt.Printf("cannot register user: %s\n", err)
				break
			}

			fmt.Println("user registration is successful")
		case `login`:
			login := readLine(`login`)
			password := readLine(`password`)

			err := sender.Login(login, password)
			if err != nil {
				fmt.Printf("cannot login user: %s\n", err)
				break
			}

			fmt.Println("user login is successful")
		case `add_data`:
			identifier := readLine(`data identifier`)
			data := readLine(`data`)

			err := sender.AddNewData(identifier, []byte(data))
			if err != nil {
				fmt.Printf("cannot add new user data: %s\n", err)
				break
			}

			fmt.Println("user data adding successful")
		case `get_data`:
			identifier := readLine(`data identifier`)

			data, err := sender.GetUserData(identifier)
			if err != nil {
				fmt.Printf("cannot get user data: %s\n", err)
				break
			}

			fmt.Println(string(data))
		}
	}
}
