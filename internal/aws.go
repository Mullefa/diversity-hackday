package internal

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

func NewSession() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-west-1"),
		},
		Profile: "developerPlayground",
	})
}

func NewRekognition() (*rekognition.Rekognition, error) {
	sess, err := NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to create new session: %w", err)
	}
	return rekognition.New(sess), nil
}

func AnalyseFace(rek *rekognition.Rekognition, filename string) (*rekognition.DetectFacesOutput, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %w", filename, err)
	}

	input := &rekognition.DetectFacesInput{
		Attributes: []*string{
			aws.String("ALL"),
		},
		Image: &rekognition.Image{
			Bytes: bytes,
		},
	}
	return rek.DetectFaces(input)
}
