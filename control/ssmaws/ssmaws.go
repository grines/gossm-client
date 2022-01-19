package ssmaws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/grines/ssmmmm-client/register"
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

func UploadSendCommand(sess *session.Session, command string, instanceID string, filename string, filedata string, currCnt int, piecesCnt int) string {

	cnt := strconv.Itoa(piecesCnt)
	ccnt := strconv.Itoa(currCnt)

	rawdata := fmt.Sprintf("%s:%s:%v:%v", ccnt, cnt, filename, filedata)

	param := make(map[string][]*string)
	param["commands"] = []*string{
		aws.String("upload " + filename),
	}

	svc := ssm.New(sess)
	input := &ssm.SendCommandInput{
		Comment:      aws.String(cnt),
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds: []*string{
			aws.String(instanceID),
		},
		OutputS3KeyPrefix: &rawdata,
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

func CreateRegistration() {
	var err error
	var key register.RsaKey

	key, err = register.CreateKeypair()

	encodedKey, err := key.EncodePrivateKey()

	//key2, err = DecodePrivateKey(encodedKey)

	encodedPublicKey, err := key.EncodePublicKey()

	fmt.Println(encodedPublicKey)
	fmt.Println()
	fmt.Println(encodedKey)
	SendRegister()
	if err != nil {
		fmt.Println("Something went wrong")
	}
}

func SendRegister() {
	httpposturl := "https://ssm.us-east-1.amazonaws.com/"
	fmt.Println("HTTP JSON POST URL:", httpposturl)

	var jsonData = []byte(`{"ActivationCode":"WC3ES5UMAtu3T80VkpV1","ActivationId":"7f8985b1-843d-4e5c-b680-b873cbffead5","Fingerprint":"b1b60ce28ceb47e7943ec7733385a33d","PublicKey":"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtRIFIEDAEi/PbNZDukTyK3XL1YYijsnLnXDJGEFiRfc/4tdjUWsJ2GC26hUpvm7gkP4JHzr6kZHrMft3NytszamgWXiqBnUhcHHrHEr7da+ewBuRKYV+6UHnQSQrFHzrvJAd6V2a0aMQcVM9WBAbmTmnDyyteCDLE6EZ/iSGgtwCBQmeplQKBJgbMsGtd4WADnj0RdDOjBzYd2jbNAiEPoJJq+Gdu+iiGD7qdtsYMrQO8OAUIPrpjYeMrOg3PznQNI8Vr1ui3CshKerdXQGgbaBXq8a4G7cDrE2ei7RdWaqag/5G2kT5/EkbARviIMsSVZ4hBoyIrazjTqlPYneUKQIDAQAB","PublicKeyType":"Rsa"}`)
	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))

	request.Header.Set("User-Agent", "aws-sdk-go/1.41.4 (go1.17.6; darwin; amd64) amazon-ssm-agent/3.1.0.0")
	request.Header.Set("X-Amz-Target", "AmazonSSM.RegisterManagedInstance")
	request.Header.Set("Content-Type", "application/x-amz-json-1.1")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

}

func GetInstanceInformation(sess *session.Session, instanceid string) *ssm.DescribeInstanceInformationOutput {

	s := make([]string, 1)
	s[0] = instanceid
	svc := ssm.New(sess)
	input := &ssm.DescribeInstanceInformationInput{
		Filters: []*ssm.InstanceInformationStringFilter{{
			Key:    aws.String("InstanceIds"),
			Values: aws.StringSlice(s),
		}},
	}

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

	return result
}

func CancelSendCommand(sess *session.Session, instanceid string, cmdid string) *ssm.CancelCommandOutput {

	s := make([]string, 1)
	s[0] = instanceid
	svc := ssm.New(sess)
	input := &ssm.CancelCommandInput{
		InstanceIds: aws.StringSlice(s),
		CommandId:   aws.String(cmdid),
	}

	result, err := svc.CancelCommand(input)
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

func GetWorkingDirectory(sess *session.Session, instanceid string) string {
	data := GetInstanceInformation(sess, instanceid)
	var wd []string

	for _, v := range data.InstanceInformationList {
		wd = append(wd, *v.PlatformName)
	}

	return wd[0]
}
