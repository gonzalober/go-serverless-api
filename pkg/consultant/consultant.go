package consultant

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/go-serverless-api/pkg/validators"
	"github.com/gofrs/uuid"
)

var (
	ErrorFailedToFetchRecord          = "failed to fetch"
	ErrorInvalidConsultantInfo        = "failed to unmarshall the object"
	ErrorInvalidConsultantData        = "invalid consultant data"
	ErrorCouldNotDeleteConsultant     = "the consultant wasn't deleted"
	ErrorCouldNotSaveConsultant       = "the consultant wasn't updated"
	ErrorConsultantAlreadyExists      = "consultant is already registered"
	ErrorConsultanDoesNotExist        = "consultant is not registered"
	ErrorInvalidEmail                 = "invalid email"
	ErrorFailedToMarshall             = "faild to marshall the object"
	ErrorCouldNotSaveConsultantDynamo = "could not save the object in dynamo"
)

type Consultant struct {
	ID              uuid.UUID      `json:"id"`
	FirstName       string         `json:"firstName"`
	LastName        string         `json:"lastName"`
	Role            string         `json:"role"`
	Since           time.Time      `json:"since"`
	Skills          []Skill        `json:"skills"`
	PastProjects    []PastProject  `json:"pastProjects"`
	Certifications  []Certificate  `json:"certifications"`
	LastModified    time.Time      `json:"lastModified"`
	Location        string         `json:"location"`
	Links           []Link         `json:"links"`
	DesiredSkills   []DesiredSkill `json:"desiredSkills"`
	ProfilePic      string         `json:"profilePic"`
	ContactInfo     []Contact      `json:"contactInfo"`
	Email           string         `json:"email"`
	PreferWFH       bool           `json:"preferWFH"`
	CurrentEmployee bool           `json:"currentEmployee"`
	CurrentStatus   string         `json:"currentStatus"`
}
type Skill struct {
	Skill         string `json:"skill"`
	Level         int    `json:"level"`
	CommercialExp bool   `json:"commercialExp"`
}

type PastProject struct {
	ProjectName    string        `json:"projectName"`
	Role           string        `json:"role"`
	Client         string        `json:"client"`
	Sector         string        `json:"sector"`
	Description    string        `json:"description"`
	StartDate      time.Time     `json:"startDate"`
	CompletionDate time.Time     `json:"completionDate"`
	Duration       time.Duration `json:"duration"`
}

type Certificate struct {
	CertificateName string        `json:"certificateName"`
	Provider        string        `json:"provider"`
	Link            url.URL       `json:"link"`
	DateAchieved    time.Time     `json:"dateAchieved"`
	Duration        time.Duration `json:"duration"`
}

type Link struct {
	LinkName string  `json:"linkName"`
	Url      url.URL `json:"url"`
}

type DesiredSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Contact struct {
	Email string `json:"email"`
	Slack string `json:"slack"`
}

func FetchConsultant(email, tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*Consultant, error) {
	// query based on email as example
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dynamoClient.GetItem(input)

	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(Consultant)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	return item, nil
}

func FetchConsultants(tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*[]Contact, error) {
	// query based on email as example
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	// not good practice do scan of all the DB
	result, err := dynamoClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	item := new([]Contact)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	return item, nil
}

func DeleteConsultant(req events.APIGatewayProxyRequest, tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*string, error) {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}
	resp, err := dynamoClient.DeleteItem(input)
	if err != nil {
		return nil, errors.New(ErrorConsultanDoesNotExist)
	}

	response := resp.String()
	fmt.Sprintf("Deleted %s", response)
	return &response, nil
}

func UpdateConsultant(req events.APIGatewayProxyRequest, tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*Consultant, error) {

	var c Consultant
	if err := json.Unmarshal([]byte(req.Body), &c); err != nil {
		return nil, errors.New(ErrorInvalidEmail)
	}

	consul, _ := FetchConsultant(c.Email, tableName, dynamoClient)

	if consul != nil && len(consul.Email) == 0 {
		return nil, errors.New(ErrorConsultanDoesNotExist)
	}

	av, err := dynamodbattribute.MarshalMap(c)

	if err != nil {
		return nil, errors.New(ErrorFailedToMarshall)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = dynamoClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotSaveConsultantDynamo)
	}
	return &c, nil
}

func CreateConsultant(req events.APIGatewayProxyRequest, tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*Consultant, error) {
	var c Consultant

	if err := json.Unmarshal([]byte(req.Body), &c); err != nil {
		return nil, errors.New(ErrorInvalidConsultantInfo)
	}

	if !validators.IsEmaiValid(c.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentConsultant, _ := FetchConsultant(c.Email, tableName, dynamoClient)
	if currentConsultant != nil && len(currentConsultant.Email) != 0 {
		return nil, errors.New(ErrorConsultantAlreadyExists)
	}
	av, err := dynamodbattribute.MarshalMap(c)
	if err != nil {
		return nil, errors.New(ErrorFailedToMarshall)
	}
	input :=
		&dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

	_, err = dynamoClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotSaveConsultantDynamo)
	}

	return &c, nil

}
