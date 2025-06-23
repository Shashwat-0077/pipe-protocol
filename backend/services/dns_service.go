package services

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

// ServerInfo represents a server with its connection details
type ServerInfo struct {
	Host     string
	Port     uint16
	Priority uint16
	Weight   uint16
	Source   string // "SRV", "MX", "A"
}

// ConnectivityResult represents the result of a connectivity test
type ConnectivityResult struct {
	Server    ServerInfo
	Connected bool
	Latency   time.Duration
	Error     error
}

// ServerResolver handles server discovery and connectivity testing
type ServerResolver struct {
	Timeout       time.Duration
	DNSTimeout    time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// NewServerResolver creates a new resolver with default settings
func NewServerResolver() *ServerResolver {
	return &ServerResolver{
		Timeout:       5 * time.Second,
		DNSTimeout:    3 * time.Second,
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}
}

var PIPE_PORTS = []uint16{9991, 587, 25, 2525, 465} // Common PIPE ports

// LookupPIPESRVRecords looks up SRV records for PIPE protocol
func (sr *ServerResolver) LookupPIPESRVRecords(domain string) ([]ServerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sr.DNSTimeout)
	defer cancel()

	// Look for _pipe._tcp.domain SRV records

	_, srvRecords, err := net.DefaultResolver.LookupSRV(ctx, "pipe", "tcp", domain)
	if err != nil {
		return nil, fmt.Errorf("PIPE SRV lookup failed for %s: %w", domain, err)
	}

	var servers []ServerInfo
	for _, srv := range srvRecords {
		servers = append(servers, ServerInfo{
			Host:     strings.TrimSuffix(srv.Target, "."),
			Port:     srv.Port,
			Priority: srv.Priority,
			Weight:   srv.Weight,
			Source:   "SRV",
		})
	}

	// Sort by priority (lower number = higher priority), then by weight (higher = preferred)
	sort.Slice(servers, func(i, j int) bool {
		if servers[i].Priority != servers[j].Priority {
			return servers[i].Priority < servers[j].Priority
		}
		return servers[i].Weight > servers[j].Weight
	})

	return servers, nil
}

// LookupMXRecords looks up MX records as fallback for email-based protocols
func (sr *ServerResolver) LookupMXRecords(domain string) ([]ServerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sr.DNSTimeout)
	defer cancel()

	mxRecords, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("MX lookup failed for %s: %w", domain, err)
	}

	var servers []ServerInfo

	for _, mx := range mxRecords {
		for _, port := range PIPE_PORTS {
			servers = append(servers, ServerInfo{
				Host:     strings.TrimSuffix(mx.Host, "."),
				Port:     port,
				Priority: mx.Pref,
				Weight:   0,
				Source:   "MX",
			})
		}
	}

	// Sort by MX priority
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Priority < servers[j].Priority
	})

	return servers, nil
}

// LookupARecords looks up A records as final fallback
func (sr *ServerResolver) LookupARecords(domain string) ([]ServerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sr.DNSTimeout)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("a record lookup failed for %s: %w", domain, err)
	}

	var servers []ServerInfo

	for _, ip := range ips {
		if ip.IP.To4() != nil { // IPv4 only
			for _, port := range PIPE_PORTS {
				servers = append(servers, ServerInfo{
					Host:     ip.IP.String(),
					Port:     port,
					Priority: 100, // Lower priority than SRV/MX
					Weight:   0,
					Source:   "A",
				})
			}
		}
	}

	return servers, nil
}

// TestTCPConnectivity tests TCP connectivity to a server
func (sr *ServerResolver) TestTCPConnectivity(server ServerInfo) ConnectivityResult {
	result := ConnectivityResult{
		Server:    server,
		Connected: false,
	}

	address := net.JoinHostPort(server.Host, fmt.Sprintf("%d", server.Port))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, sr.Timeout)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	conn.Close()
	result.Connected = true
	return result
}

// TestTCPConnectivityWithRetries tests TCP connectivity with retries
func (sr *ServerResolver) TestTCPConnectivityWithRetries(server ServerInfo) ConnectivityResult {
	var lastResult ConnectivityResult

	for attempt := 0; attempt < sr.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(sr.RetryInterval)
		}

		result := sr.TestTCPConnectivity(server)
		if result.Connected {
			return result
		}
		lastResult = result
	}

	return lastResult
}

// DiscoverPipeServers discovers PIPE servers for a domain using multiple methods
func (sr *ServerResolver) DiscoverPipeServers(domain string) ([]ServerInfo, error) {
	var allServers []ServerInfo

	// 1. Try PIPE SRV records first (highest priority)
	if srvServers, err := sr.LookupPIPESRVRecords(domain); err == nil {
		allServers = append(allServers, srvServers...)
	}

	// 2. Try MX records as fallback
	if mxServers, err := sr.LookupMXRecords(domain); err == nil {
		allServers = append(allServers, mxServers...)
	}

	// 3. Try A records as final fallback
	if aServers, err := sr.LookupARecords(domain); err == nil {
		allServers = append(allServers, aServers...)
	}

	if len(allServers) == 0 {
		return nil, fmt.Errorf("no PIPE servers found for domain: %s", domain)
	}

	// Sort all servers by priority, then source preference
	sort.Slice(allServers, func(i, j int) bool {
		// SRV records have highest priority
		if allServers[i].Source != allServers[j].Source {
			sourceOrder := map[string]int{"SRV": 0, "MX": 1, "A": 2}
			return sourceOrder[allServers[i].Source] < sourceOrder[allServers[j].Source]
		}
		// Within same source, sort by priority
		return allServers[i].Priority < allServers[j].Priority
	})

	return allServers, nil
}

// FindWorkingPipeServers discovers and tests PIPE servers, returning results sorted by connectivity
func (sr *ServerResolver) FindWorkingPipeServers(domain string) ([]ConnectivityResult, error) {
	servers, err := sr.DiscoverPipeServers(domain)
	if err != nil {
		return nil, err
	}

	var results []ConnectivityResult
	for _, server := range servers {
		result := sr.TestTCPConnectivityWithRetries(server)
		results = append(results, result)
	}

	// Sort results: working servers first, then by priority/latency
	sort.Slice(results, func(i, j int) bool {
		if results[i].Connected != results[j].Connected {
			return results[i].Connected // Connected servers first
		}
		if results[i].Server.Priority != results[j].Server.Priority {
			return results[i].Server.Priority < results[j].Server.Priority
		}
		return results[i].Latency < results[j].Latency
	})

	return results, nil
}

// GetBestPipeServer returns the best working PIPE server for a domain
func (sr *ServerResolver) GetBestPipeServer(domain string) (*ConnectivityResult, error) {
	results, err := sr.FindWorkingPipeServers(domain)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		if result.Connected {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("no working PIPE servers found for domain: %s", domain)
}

// PrintServerStatus prints a formatted status of all servers
func PrintServerStatus(results []ConnectivityResult) {
	fmt.Println("PIPE Server Status Report:")
	fmt.Println("==========================")

	for i, result := range results {
		status := "❌ FAILED"
		if result.Connected {
			status = "✅ CONNECTED"
		}

		fmt.Printf("%d. %s:%d [%s] %s\n",
			i+1,
			result.Server.Host,
			result.Server.Port,
			result.Server.Source,
			status,
		)

		if result.Connected {
			fmt.Printf("   Priority: %d, Weight: %d, Latency: %v\n",
				result.Server.Priority,
				result.Server.Weight,
				result.Latency,
			)
		} else if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
		fmt.Println()
	}
}

// Quick helper functions for common operations

// QuickConnect attempts to connect to the best available PIPE server
func QuickConnect(domain string) (*ConnectivityResult, error) {
	resolver := NewServerResolver()
	return resolver.GetBestPipeServer(domain)
}

// QuickScan scans and returns all PIPE server statuses for a domain
func QuickScan(domain string) ([]ConnectivityResult, error) {
	resolver := NewServerResolver()
	return resolver.FindWorkingPipeServers(domain)
}

// GetPipeServerList returns just the discovered servers without testing connectivity
func GetPipeServerList(domain string) ([]ServerInfo, error) {
	resolver := NewServerResolver()
	return resolver.DiscoverPipeServers(domain)
}
