package email

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESConfig configures the ses driver.
//
// SES also speaks SMTP, so the smtp driver works with it too. The API driver exists for one
// reason: on AWS it can authenticate with the instance's IAM role (or IRSA), leaving no SMTP
// password to store and rotate.
type SESConfig struct {
	// Region is the SES region the sender identity is verified in. Empty uses the default
	// chain (AWS_REGION, shared config).
	Region string `mapstructure:"region"`
	// AccessKeyID and SecretAccessKey are static credentials. Leave both empty to use the
	// default AWS credential chain — the right choice on AWS.
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	// ConfigurationSet names an SES configuration set, for event publishing and dedicated IPs.
	ConfigurationSet string `mapstructure:"configuration_set"`
	// Endpoint overrides the service endpoint. Empty means AWS; set it for LocalStack.
	Endpoint string `mapstructure:"endpoint"`
	// Timeout bounds one request. Default 30s.
	Timeout time.Duration `mapstructure:"timeout"`
}

// SESSender delivers through the Amazon SES v2 API.
type SESSender struct {
	from   *mail.Address
	cfg    SESConfig
	client *sesv2.Client
}

// NewSESSender loads AWS configuration and returns a sender. ctx is used only to load it.
func NewSESSender(ctx context.Context, from *mail.Address, cfg SESConfig) (*SESSender, error) {
	var loadOptions []func(*awsconfig.LoadOptions) error
	if cfg.Region != "" {
		loadOptions = append(loadOptions, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := sesv2.NewFromConfig(awsCfg, func(o *sesv2.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(strings.TrimSuffix(cfg.Endpoint, "/"))
		}
	})

	return &SESSender{from: from, cfg: cfg, client: client}, nil
}

// Send delivers msg.
func (s *SESSender) Send(ctx context.Context, msg Message) error {
	to, err := msg.validate()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutOrDefault(s.cfg.Timeout))
	defer cancel()

	body := &types.Body{Text: &types.Content{Data: aws.String(msg.Text), Charset: aws.String("UTF-8")}}
	if msg.HTML != "" {
		body.Html = &types.Content{Data: aws.String(msg.HTML), Charset: aws.String("UTF-8")}
	}

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(formatAddress(s.from)),
		Destination:      &types.Destination{ToAddresses: []string{formatAddress(to)}},
		Content: &types.EmailContent{Simple: &types.Message{
			Subject: &types.Content{Data: aws.String(msg.Subject), Charset: aws.String("UTF-8")},
			Body:    body,
		}},
	}
	if s.cfg.ConfigurationSet != "" {
		input.ConfigurationSetName = aws.String(s.cfg.ConfigurationSet)
	}

	if _, err := s.client.SendEmail(ctx, input); err != nil {
		return fmt.Errorf("ses: %w", err)
	}
	return nil
}
