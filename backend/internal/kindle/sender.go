package kindle

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// SupportedFormats lists formats that Kindle supports
var SupportedFormats = []string{".mobi", ".azw", ".azw3", ".pdf", ".epub", ".doc", ".docx", ".txt", ".rtf", ".html"}

// KindleSender handles sending books to Kindle devices
type KindleSender struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	useTLS       bool
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// NewKindleSender creates a new Kindle sender
func NewKindleSender(config SMTPConfig) *KindleSender {
	return &KindleSender{
		smtpHost:     config.Host,
		smtpPort:     config.Port,
		smtpUsername: config.Username,
		smtpPassword: config.Password,
		fromEmail:    config.From,
		useTLS:       config.UseTLS,
	}
}

// SendResult represents the result of a send operation
type SendResult struct {
	Success  bool
	Error    string
	FilePath string
	ToEmail  string
}

// Send sends a book file to a Kindle email address
func (k *KindleSender) Send(filePath, kindleEmail string) (*SendResult, error) {
	result := &SendResult{
		FilePath: filePath,
		ToEmail:  kindleEmail,
	}

	// Validate file exists
	if _, err := os.Stat(filePath); err != nil {
		result.Error = fmt.Sprintf("file not found: %s", filePath)
		return result, fmt.Errorf(result.Error)
	}

	// Validate format
	ext := strings.ToLower(filepath.Ext(filePath))
	supported := false
	for _, f := range SupportedFormats {
		if ext == f {
			supported = true
			break
		}
	}
	if !supported {
		result.Error = fmt.Sprintf("unsupported format: %s", ext)
		return result, fmt.Errorf(result.Error)
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read file: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	// Build email
	fileName := filepath.Base(filePath)
	subject := "Kindle Document: " + strings.TrimSuffix(fileName, ext)
	
	message, err := k.buildEmailWithAttachment(kindleEmail, subject, fileName, fileData)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build email: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	// Send email
	if err := k.sendEmail(kindleEmail, message); err != nil {
		result.Error = fmt.Sprintf("failed to send email: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	result.Success = true
	return result, nil
}

// SendWithConversion sends a book, converting if necessary
func (k *KindleSender) SendWithConversion(filePath, kindleEmail string, converter EbookConverter) (*SendResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Check if conversion is needed
	// EPUB needs to be converted to MOBI for older Kindles, but modern Kindles support EPUB
	// For best compatibility, we'll allow EPUB to be sent directly since newer Kindles support it
	if ext == ".epub" || ext == ".mobi" || ext == ".azw3" || ext == ".pdf" {
		return k.Send(filePath, kindleEmail)
	}

	// Convert other formats to MOBI
	if converter != nil && converter.IsAvailable() {
		outputPath := strings.TrimSuffix(filePath, ext) + ".mobi"
		
		// Perform conversion
		ctx := converter.Context()
		if _, err := converter.ConvertToFormat(ctx, filePath, "mobi", nil); err == nil {
			// Send converted file
			result, sendErr := k.Send(outputPath, kindleEmail)
			
			// Clean up converted file
			os.Remove(outputPath)
			
			return result, sendErr
		}
	}

	// Fall back to sending original
	return k.Send(filePath, kindleEmail)
}

func (k *KindleSender) buildEmailWithAttachment(to, subject, filename string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Headers
	headers := make(textproto.MIMEHeader)
	headers.Set("From", k.fromEmail)
	headers.Set("To", to)
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))
	
	// Write headers
	for key, values := range headers {
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buf.WriteString("\r\n")
	
	// Empty body part
	textPart, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	})
	textPart.Write([]byte("Sent from Shelfarr\r\n"))
	
	// Attachment part
	contentType := k.getContentType(filename)
	attachmentHeader := textproto.MIMEHeader{
		"Content-Type":              []string{contentType + "; name=\"" + filename + "\""},
		"Content-Transfer-Encoding": []string{"base64"},
		"Content-Disposition":       []string{"attachment; filename=\"" + filename + "\""},
	}
	
	attachmentPart, err := writer.CreatePart(attachmentHeader)
	if err != nil {
		return nil, err
	}
	
	// Write base64 encoded data with line breaks
	encoded := base64.StdEncoding.EncodeToString(data)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		attachmentPart.Write([]byte(encoded[i:end] + "\r\n"))
	}
	
	writer.Close()
	
	return buf.Bytes(), nil
}

func (k *KindleSender) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	contentTypes := map[string]string{
		".mobi": "application/x-mobipocket-ebook",
		".azw":  "application/vnd.amazon.ebook",
		".azw3": "application/vnd.amazon.ebook",
		".epub": "application/epub+zip",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".rtf":  "application/rtf",
		".html": "text/html",
	}
	
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func (k *KindleSender) sendEmail(to string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", k.smtpHost, k.smtpPort)
	
	var client *smtp.Client
	var err error
	
	if k.useTLS {
		// Connect with TLS
		tlsConfig := &tls.Config{
			ServerName: k.smtpHost,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		
		client, err = smtp.NewClient(conn, k.smtpHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
	} else {
		// Connect without TLS, possibly upgrade with STARTTLS
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		
		client, err = smtp.NewClient(conn, k.smtpHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
		
		// Try STARTTLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: k.smtpHost,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				// Continue without TLS if STARTTLS fails
			}
		}
	}
	defer client.Close()
	
	// Authenticate if credentials provided
	if k.smtpUsername != "" && k.smtpPassword != "" {
		auth := smtp.PlainAuth("", k.smtpUsername, k.smtpPassword, k.smtpHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}
	
	// Set sender
	if err := client.Mail(k.fromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	
	// Set recipient
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	
	// Send data
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}
	
	return client.Quit()
}

// Test tests the SMTP connection
func (k *KindleSender) Test() error {
	addr := fmt.Sprintf("%s:%d", k.smtpHost, k.smtpPort)
	
	var conn net.Conn
	var err error
	
	if k.useTLS {
		tlsConfig := &tls.Config{
			ServerName: k.smtpHost,
		}
		conn, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()
	
	client, err := smtp.NewClient(conn, k.smtpHost)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()
	
	// Try STARTTLS if not using TLS
	if !k.useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: k.smtpHost,
			}
			client.StartTLS(tlsConfig)
		}
	}
	
	// Test authentication
	if k.smtpUsername != "" && k.smtpPassword != "" {
		auth := smtp.PlainAuth("", k.smtpUsername, k.smtpPassword, k.smtpHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}
	
	return client.Quit()
}

// EbookConverter interface for optional format conversion
type EbookConverter interface {
	IsAvailable() bool
	Context() interface{}
	ConvertToFormat(ctx interface{}, inputPath, outputFormat string, opts interface{}) (interface{}, error)
}

// ValidateKindleEmail validates a Kindle email address
func ValidateKindleEmail(email string) bool {
	// Kindle emails must end with @kindle.com or @free.kindle.com
	email = strings.ToLower(email)
	return strings.HasSuffix(email, "@kindle.com") || 
		   strings.HasSuffix(email, "@free.kindle.com") ||
		   strings.HasSuffix(email, "@kindle.cn")
}

// SendQueue handles queuing and batch sending
type SendQueue struct {
	sender *KindleSender
	queue  []SendRequest
}

// SendRequest represents a queued send request
type SendRequest struct {
	FilePath    string
	KindleEmail string
	BookID      uint
}

// NewSendQueue creates a new send queue
func NewSendQueue(sender *KindleSender) *SendQueue {
	return &SendQueue{
		sender: sender,
		queue:  make([]SendRequest, 0),
	}
}

// Add adds a send request to the queue
func (q *SendQueue) Add(req SendRequest) {
	q.queue = append(q.queue, req)
}

// Process processes all queued sends
func (q *SendQueue) Process() []SendResult {
	results := make([]SendResult, 0, len(q.queue))
	
	for _, req := range q.queue {
		result, _ := q.sender.Send(req.FilePath, req.KindleEmail)
		results = append(results, *result)
	}
	
	// Clear queue
	q.queue = q.queue[:0]
	
	return results
}

// Len returns the queue length
func (q *SendQueue) Len() int {
	return len(q.queue)
}

// Compile-time check that we implement io.Writer
var _ io.Writer = (*bytes.Buffer)(nil)

