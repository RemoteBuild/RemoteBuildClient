package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/JojiiOfficial/configService"
	"github.com/JojiiOfficial/gaw"
	"github.com/zalando/go-keyring"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

// LoginCommand login into the server
func (cData *CommandData) LoginCommand(usernameArg string, args ...bool) {
	// Print confirmation if user is already logged in
	if cData.Config.IsLoggedIn() && !cData.Yes && len(args) == 0 {
		i, _ := gaw.ConfirmInput("You are already logged in. Overwrite session? [y/n]> ", bufio.NewReader(os.Stdin))
		if !i {
			return
		}
	}

	// Enter credentials
	username, pass := credentials(usernameArg, false, 0)

	// Do HTTP request
	loginResponse, err := cData.Librb.Login(username, pass)
	if err != nil {
		printResponseError(err, "logging in")
		return
	}

	cData.Config.InsertUser(username, loginResponse.Token)

	// Save new config
	err = configService.Save(cData.Config, cData.Config.File)
	if err != nil {
		fmtError("saving config:", err.Error())
		return
	}

	fmt.Println(color.HiGreenString("Success!"), "\nLogged in as", username)
}

// RegisterCommand create a new account
func (cData *CommandData) RegisterCommand() {
	// Input for credentials
	username, pass := credentials("", true, 0)
	if len(username) == 0 || len(pass) == 0 {
		return
	}

	// Do HTTP request
	_, err := cData.Librb.Register(username, pass)
	if err != nil {
		printResponseError(err, "creating an account")
		return
	}

	fmt.Printf("User '%s' created %s!\n", username, color.HiGreenString("successfully"))

	// Ask for login
	y, _ := gaw.ConfirmInput("Do you want to login to this account? [y/n]> ", bufio.NewReader(os.Stdin))
	if y {
		cData.LoginCommand(username, true)
	}
}

// Logout Logs out the user
func (cData *CommandData) Logout(username string) {
	err := cData.Config.ClearKeyring(username)
	if err == nil || err == keyring.ErrNotFound {
		printSuccess("logged out")
	} else {
		fmt.Println(err)
	}
}

func credentials(bUser string, repeat bool, index uint8) (string, string) {
	if index >= 3 {
		return "", ""
	}

	reader := bufio.NewReader(os.Stdin)
	var username string
	if len(bUser) > 0 {
		username = bUser
	} else {
		fmt.Print("Enter Username: ")
		username, _ = reader.ReadString('\n')
	}
	username = strings.ToLower(username)

	if len(username) > 30 {
		fmt.Println("Username too long!")
		return "", ""
	}

	var pass string
	enterPassMsg := "Enter Password: "
	count := 1

	if repeat {
		count = 2
	}

	for i := 0; i < count; i++ {
		fmt.Print(enterPassMsg)
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalln("Error:", err.Error())
			return "", ""
		}
		fmt.Println()
		lPass := strings.TrimSpace(string(bytePassword))

		if len(lPass) > 80 {
			fmt.Println("Your password is too long!")
			return credentials(username, repeat, index+1)
		}
		if len(lPass) < 7 {
			fmt.Println("Your password must have at least 7 characters!")
			return credentials(username, repeat, index+1)
		}

		if repeat && i == 1 && pass != lPass {
			fmt.Println("Passwords don't match!")
			return credentials(username, repeat, index+1)
		}

		pass = lPass
		enterPassMsg = "Enter Password again: "
	}

	return strings.TrimSpace(username), pass
}

// Ping pings the server
func Ping(cData *CommandData) {
	pingResp, err := cData.Librb.Ping()
	if err != nil {
		printError("while pinging", err.Error())
		return
	}
	fmt.Println(pingResp.String)
}
