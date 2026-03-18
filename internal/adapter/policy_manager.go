package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var HOSTNAME_REGEX = regexp.MustCompile(
	`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])$`,
)

func NewPolicyManager(
	storageSvcHostname,
	storageSvcAccessKey,
	storageSvcPolicyEndpoint string,
	storageSvcPort uint16,
) *PolicyManager {

	return &PolicyManager{
		storageSvcHostname:       storageSvcHostname,
		storageSvcAccessKey:      storageSvcAccessKey,
		storageSvcPort:           storageSvcPort,
		storageSvcPolicyEndpoint: storageSvcPolicyEndpoint,
		policies:                 make(map[string]*UploadPolicy),
	}
}

func (pm *PolicyManager) GetPolicy(ctx context.Context, jobID string) (*UploadPolicy, error) {
	pm.mu.RLock()
	policy, exists := pm.policies[jobID]
	pm.mu.RUnlock()

	// If policy is in store, return it
	if exists {
		return policy, nil
	}

	// Acquire a Write Lock so other concurrent ffmpeg requests wait
	// while we fetch the new policy instead of spamming Spring Boot.
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Double-check inside the lock in case another goroutine JUST fetched it
	policy, exists = pm.policies[jobID]
	if exists {
		return policy, nil
	}

	log.Printf("Fetching fresh MinIO policy for job ID: %s\n", jobID)

	policy, err := fetchUploadPolicy(ctx, pm.storageSvcHostname, pm.storageSvcPolicyEndpoint, pm.storageSvcPort, pm.storageSvcAccessKey, false)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	pm.policies[jobID] = policy
	return policy, nil
}

func fetchUploadPolicy(
	ctx context.Context,
	storageSvcHostname,
	policyGenerationEndpoint string,
	storageSvcPort uint16,
	accessKey string,
	https bool,
) (*UploadPolicy, error) {
	protocol := "http"
	if https {
		protocol = protocol + "s"
	}

	newPolicyGenerationEndpoint, _ := strings.CutPrefix(policyGenerationEndpoint, "/")

	if !HOSTNAME_REGEX.Match([]byte(storageSvcHostname)) {
		return nil, fmt.Errorf("error: invalid hostname")
	}

	u := url.URL{
		Scheme: protocol,
		Host:   fmt.Sprintf("%s:%d", storageSvcHostname, storageSvcPort),
		Path:   newPolicyGenerationEndpoint,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	
	var policyData UploadPolicy

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("storage service returned status: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&policyData); err != nil {
		return nil, fmt.Errorf("failed to decode policy: %w", err)
	}

	return &policyData, nil
}
