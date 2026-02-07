package mail

import (
	"github.com/knadh/go-pop3"
)

type MailHandlingService struct {
	client *pop3.Client
}

func NewMailHandlingService() *MailHandlingService {
	return &MailHandlingService{}
}
