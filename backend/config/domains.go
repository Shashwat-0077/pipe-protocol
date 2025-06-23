package config

import (
	"os"
	"strings"
)

type Config struct {
	LocalDomains []string
	ServerDomain string
	PipePort     int
}

var AppConfig *Config

func init() {
	AppConfig = &Config{
		LocalDomains: getLocalDomains(),
		ServerDomain: getServerDomain(),
		PipePort:     getPipePort(),
	}
}

func getLocalDomains() []string {
	domains := os.Getenv("LOCAL_DOMAINS")
	if domains == "" {
		return []string{"pipec.local", "localhost"}
	}
	return strings.Split(domains, ",")
}

func getServerDomain() string {
	domain := os.Getenv("SERVER_DOMAIN")
	if domain == "" {
		return "pipec.local"
	}
	return domain
}

func getPipePort() int {
	// You can add port parsing logic here
	return 2525 // Default PIPE port
}

// IsLocalDomain checks if domain is local
func (c *Config) IsLocalDomain(domain string) bool {
	for _, localDomain := range c.LocalDomains {
		if strings.EqualFold(domain, localDomain) {
			return true
		}
	}
	return false
}
