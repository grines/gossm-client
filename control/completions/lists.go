package completion

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func listProfiles() func(string) []string {
	return func(line string) []string {
		rule := `\[(.*)\]`
		var profiles []string

		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		dat, err := ioutil.ReadFile(usr.HomeDir + "/.aws/credentials")
		if err != nil {
			fmt.Println(err)
		}

		r, _ := regexp.Compile(rule)
		if r.MatchString(string(dat)) {
			matches := r.FindAllStringSubmatch(string(dat), -1)
			for _, v := range matches {
				profiles = append(profiles, v[1])
			}
		}
		return profiles
	}
}

func listInstances(sess *session.Session) func(string) []string {
	return func(line string) []string {
		var instances []string

		svc := ssm.New(sess)
		input := &ssm.DescribeInstanceInformationInput{}

		result, err := svc.DescribeInstanceInformation(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return nil
		}
		for _, v := range result.InstanceInformationList {
			if v.ComputerName != nil && v.InstanceId != nil {
				instances = append(instances, *v.ComputerName+":"+*v.InstanceId)
			}

		}

		return instances
	}
}
