package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	URL := ""
	if https {
		protocol = protocol + "s"
	}

	newPolicyGenerationEndpoint, _ := strings.CutPrefix(policyGenerationEndpoint, "/")

	URL += fmt.Sprintf(
		"%s://%s:%v/%s",
		protocol,
		storageSvcHostname,
		storageSvcPort,
		newPolicyGenerationEndpoint,
	)
	res, err := http.Get(URL)

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
