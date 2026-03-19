package adapter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

func NewService(
	storageSvcAccessKey,
	storageSvcHostname,
	storageSvcPolicyEndpoint string,
	storageSvcPort uint16,
) (*Service, error) {
	pm, err := NewPolicyManager(storageSvcHostname, storageSvcAccessKey, storageSvcPolicyEndpoint, storageSvcPort)

	if err != nil {
		return nil, err
	}
	return &Service{
		pm: pm,
	}, nil
}

func (svc *Service) MakePostRequest(
	ctx context.Context,
	jobID string,
	filePath string,
	fileStream io.Reader,
) error {
	policy, err := svc.pm.GetPolicy(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	var bodyBuf bytes.Buffer
	writer := multipart.NewWriter(&bodyBuf)

	for key, val := range policy.Fields {
		if key == "key" {
			continue
		}
		if err := writer.WriteField(key, val); err != nil {
			return fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	fullMinioKey := policy.KeyPrefix + filePath
	log.Println("minio key is: ", fullMinioKey)
	if err := writer.WriteField("key", fullMinioKey); err != nil {
		return fmt.Errorf("failed to write key field: %w", err)
	}

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, fileStream); err != nil {
		return fmt.Errorf("failed to copy file stream to buffer: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	ctxWithoutCancel := context.WithoutCancel(ctx)
	req, err := http.NewRequestWithContext(ctxWithoutCancel, http.MethodPost, policy.URL, &bodyBuf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	var resp *http.Response
	client := &http.Client{}
	for i := 0; i < 5; i++ {
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("error in minio upload. Tries (%d)\n", i+1)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			err = nil
			break
		}
	}
	if err != nil {
		return fmt.Errorf("minio request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("minio rejected upload (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
