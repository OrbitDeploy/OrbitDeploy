package services

import (
	"fmt"
	"log"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/OrbitDeploy/fastcaddy"
)

const SystemPort = 8285

// SystemDomainService handles system-level domain management.
type SystemDomainService struct{}

// NewSystemDomainService creates a new SystemDomainService.
func NewSystemDomainService() *SystemDomainService {
	return &SystemDomainService{}
}

// UpdateSystemDomain updates the system's main domain.
func (s *SystemDomainService) UpdateSystemDomain(newDomain string) error {
	// 1. Get the old domain from the database
	oldDomain, err := models.GetSystemSetting("system_domain")
	if err != nil {
		log.Printf("Error getting old system domain: %v", err)
		return fmt.Errorf("failed to get old system domain: %w", err)
	}

	// 2. Normalize the new domain
	cleanNewDomain, err := utils.NormalizeDomain(newDomain)
	if err != nil {
		log.Printf("Invalid new domain format: %v", err)
		return fmt.Errorf("invalid new domain format: %w", err)
	}

	// If the domain hasn't changed, do nothing
	if oldDomain == cleanNewDomain {
		log.Printf("System domain is already set to %s, no changes needed.", cleanNewDomain)
		return nil
	}

	fc := fastcaddy.New()

	// 3. Remove the old domain from Caddy if it exists
	if oldDomain != "" {
		log.Printf("Removing old system domain from Caddy: %s", oldDomain)
		if err := fc.DeleteRoute(oldDomain); err != nil {
			// Log the error but don't block the process, as the route might not exist
			log.Printf("Warning: Failed to delete old route %s from Caddy: %v", oldDomain, err)
		}
	}

	// 4. Add the new domain to Caddy if it's not empty
	if cleanNewDomain != "" {
		proxyTo := fmt.Sprintf("localhost:%d", SystemPort)
		log.Printf("Adding new system domain to Caddy: %s -> %s", cleanNewDomain, proxyTo)
		if err := fc.AddReverseProxy(cleanNewDomain, proxyTo); err != nil {
			log.Printf("Error adding new system domain route to Caddy: %v", err)
			// Try to roll back by re-adding the old domain if it existed
			if oldDomain != "" {
				oldProxyTo := fmt.Sprintf("localhost:%d", SystemPort)
				if rbErr := fc.AddReverseProxy(oldDomain, oldProxyTo); rbErr != nil {
					log.Printf("CRITICAL: Failed to roll back Caddy config for old domain %s: %v", oldDomain, rbErr)
				}
			}
			return fmt.Errorf("failed to add new route to Caddy: %w", err)
		}
	}

	log.Printf("Successfully updated system domain to: %s", cleanNewDomain)
	return nil
}
