package ssmaws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func DescribeInstances(sess *session.Session) []string {

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
		if v.ComputerName != nil {
			instances = append(instances, *v.ComputerName)
		}

	}

	return instances
}

func SendCommand(sess *session.Session, command string, instanceID string) string {

	param := make(map[string][]*string)
	param["commands"] = []*string{
		aws.String(command),
	}

	svc := ssm.New(sess)
	input := &ssm.SendCommandInput{
		Comment:      aws.String(command),
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	result, err := svc.SendCommand(input)
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
		return ""
	}

	return *result.Command.CommandId
}

func GetCommandOutput(sess *session.Session, commandid string, instanceid string) *ssm.GetCommandInvocationOutput {

	svc := ssm.New(sess)
	input := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandid),
		InstanceId: aws.String(instanceid),
	}

	result, err := svc.GetCommandInvocation(input)
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

	return result
}
