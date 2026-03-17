package adapter

import "sync"

type UploadPolicy struct {
	URL       string
	Fields    map[string]string
	KeyPrefix string
}

type Service struct {
	pm *PolicyManager
}

type PolicyManager struct {
	storageSvcHostname string
	accessKey          string
	port               uint16
	mu                 sync.RWMutex
	policies           map[string]*UploadPolicy
}

type Handler struct {
	svc Service
}
