package gui

import (
	"net"
	"strings"

	"go-chat/internal/network"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidateMultiaddr validates a libp2p multiaddr string.
func ValidateMultiaddr(addr string) error {
	if addr == "" {
		return &ValidationError{Field: "addr", Message: "address cannot be empty"}
	}

	// Basic multiaddr format check
	if !strings.HasPrefix(addr, "/") {
		return &ValidationError{Field: "addr", Message: "address must start with '/'"}
	}

	// Try to parse as multiaddr
	if ip := net.ParseIP(addr); ip != nil {
		// This is just an IP address, not a full multiaddr
		return &ValidationError{Field: "addr", Message: "address must be a full multiaddr (e.g. /ip4/1.2.3.4/tcp/4001/p2p/...)"}
	}

	// Additional validation could go here
	return nil
}

// ValidateDisplayName validates a display name.
func ValidateDisplayName(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "display name cannot be empty"}
	}
	if len(name) > 50 {
		return &ValidationError{Field: "name", Message: "display name too long (max 50 characters)"}
	}
	// Check for reserved names
	if name == "" || name == "me" || strings.HasPrefix(name, "me_") {
		return &ValidationError{Field: "name", Message: "display name is reserved"}
	}
	return nil
}

// ValidateChannelName validates a channel name.
func ValidateChannelName(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "channel name cannot be empty"}
	}
	if len(name) > 100 {
		return &ValidationError{Field: "name", Message: "channel name too long (max 100 characters)"}
	}
	return nil
}

// ValidateOrganizationName validates an organization name.
func ValidateOrganizationName(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "organization name cannot be empty"}
	}
	if len(name) > 100 {
		return &ValidationError{Field: "name", Message: "organization name too long (max 100 characters)"}
	}
	return nil
}

// ValidateProtocolID validates that a protocol ID is supported.
func ValidateProtocolID(id string) error {
	if id == "" {
		return &ValidationError{Field: "protocol", Message: "protocol ID cannot be empty"}
	}
	if id != string(network.ProtocolID) {
		return &ValidationError{Field: "protocol", Message: "unsupported protocol ID"}
	}
	return nil
}