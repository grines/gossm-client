package completion

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

//Load profile from .aws/credentials by name
func getProfile(pname string, region string) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", pname),
	})
	if err != nil {
		fmt.Println("Invalid Credentials")
		connected = false
	} else {
		connected = true
	}
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		fmt.Println("Invalid Credentials")
		connected = false
	} else {
		connected = true
	}
	return sess
}

func GetCallerIdentity(sess *session.Session) string {
	svc := sts.New(sess)

	var params *sts.GetCallerIdentityInput
	resp, err := svc.GetCallerIdentity(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return ""
	}

	return string(*resp.Arn)
}
