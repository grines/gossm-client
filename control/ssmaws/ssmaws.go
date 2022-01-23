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
	uuid "github.com/nu7hatch/gouuid"
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

func GetCommandOutputDocs(sess *session.Session, commandid string, instanceid string) (*ssm.GetDocumentOutput, error) {

	svc := ssm.New(sess)
	input := &ssm.GetDocumentInput{
		Name: aws.String(commandid),
	}

	result, err := svc.GetDocument(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, err
	}

	return result, err
}

func DeleteOutputDoc(sess *session.Session, commandid string, instanceid string) (*ssm.DeleteDocumentOutput, error) {

	svc := ssm.New(sess)
	input := &ssm.DeleteDocumentInput{
		Name: aws.String(commandid),
	}

	result, err := svc.DeleteDocument(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, err
	}

	return result, err
}

func CreateRegistration(activationCode string, activationID string, region string) {
	var err error
	var key register.RsaKey

	key, err = register.CreateKeypair()

	encodedKey, err := key.EncodePrivateKey()

	//key2, err = DecodePrivateKey(encodedKey)

	encodedPublicKey, err := key.EncodePublicKey()

	fmt.Println(encodedPublicKey)
	fmt.Println()
	fmt.Println(encodedKey)
	SendRegister(activationCode, activationID, encodedPublicKey, region)
	if err != nil {
		fmt.Println("Something went wrong")
	}
}

func SendRegister(activationCode string, activationID string, PublicKey string, region string) {
	httpposturl := "https://ssm." + region + ".amazonaws.com/"
	fmt.Println("HTTP JSON POST URL:", httpposturl)

	fingerprint, _ := uuid.NewV4()
	fmt.Println(fingerprint.String())

	var jsonData = []byte(`{"ActivationCode":"` + activationCode + `","ActivationId":"` + activationID + `","Fingerprint":"` + fingerprint.String() + `","PublicKey":"` + PublicKey + `","PublicKeyType":"Rsa"}`)
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
