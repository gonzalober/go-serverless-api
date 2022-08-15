package consultant

import (
	"errors"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/gofrs/uuid"
)

var (
	ErrorFailedToFetchRecord = "failed to fetch"
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

func FetchConsultants(tableName string, dynamoClient dynamodbiface.DynamoDBAPI) (*[]Consultant, error) {
	// query based on email as example
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynamoClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	item := new([]Consultant)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item)
	return item, nil
}
