package mailer

import (
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"encoding/base64"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"os"
)

type Message struct {
	ID         uuid.UUID         `json:"id"`
	UserID     string            `json:"user_id"`
	Target     string            `json:"target"`
	Type       MessageType       `json:"type"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	TemplateID string            `json:"template_id"`
	DataMap    map[string]string `json:"data_map"`
	Ts         int64             `json:"ts"`
}

// MessageType enum type
type MessageType string

const (
	PUSH_MESSAGE_TYPE  MessageType = "PUSH"
	EMAIL_MESSAGE_TYPE MessageType = "EMAIL"
	SMS_MESSAGE_TYPE   MessageType = "SMS"
)

func SendMail(data Message, templatePath string, logger *zap.Logger, mailConfig *config.MailerConfig) error {
	newRequestData := NewRequest(data.Target, data.Title, logger, mailConfig)
	go func() {
		err := newRequestData.AppSendMail(templatePath, data)
		if err != nil {
			return
		}
	}()
	return nil
}

func EncodeImages(imagePath string) string {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return ""
	}

	base64Data := base64.StdEncoding.EncodeToString(imageData)

	return base64Data
}
