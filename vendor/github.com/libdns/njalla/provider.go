// Package njalla implements a DNS record management client compatible
// with the libdns interfaces for Njalla DNS.
package njalla

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
)

// Provider implementation for Njalla service
// This provider uses sync.Once to implicitly provision the client when methods are called
// No additional provisioning steps are required by callers

// Provider facilitates DNS record manipulation with Njalla.
// It implements the libdns interfaces for DNS record management
// using the Njalla API.
//
// This implementation supports all standard record types that Njalla offers,
// with special handling for the most common types (A, AAAA, CNAME, TXT, MX, SRV).
// Other record types are handled using the generic libdns.RR type.
//
// All methods are safe for concurrent use.
type Provider struct {
	// APIToken is the Njalla API token required for authentication
	// You can generate an API token from your Njalla account settings.
	APIToken string `json:"api_token,omitempty"`

	client     clientInterface
	clientOnce sync.Once
}

// getClient lazily initializes the API client
func (p *Provider) getClient() clientInterface {
	p.clientOnce.Do(func() {
		if p.APIToken == "" {
			// Don't panic, let methods handle this gracefully
			return
		}
		p.client = newClient(p.APIToken)
	})
	return p.client
}

// GetRecords lists all the records in the zone.
// It retrieves all DNS records from the specified zone and returns them as libdns.Record types.
// This method implements the libdns.RecordGetter interface.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	if p.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	client := p.getClient()
	if client == nil {
		return nil, fmt.Errorf("API client not initialized")
	}

	// Ensure zone name is in the proper format (remove trailing dot for API calls)
	zone = strings.TrimSuffix(zone, ".")

	// List records
	req := listRecordsRequest{Domain: zone}
	var resp listRecordsResponse

	if err := client.call(ctx, "list-records", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// Convert to libdns.Record
	records := make([]libdns.Record, 0, len(resp.Records))
	for _, record := range resp.Records {
		libdnsRecord, err := njallaRecordToLibdns(record)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record: %w", err)
		}
		records = append(records, libdnsRecord)
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
// This method creates new DNS records in the specified zone without modifying existing records.
// The returned records include the provider-specific IDs in the ProviderData field,
// which can be used for future operations like updates and deletions.
//
// This method implements the libdns.RecordAppender interface.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if p.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	client := p.getClient()
	if client == nil {
		return nil, fmt.Errorf("API client not initialized")
	}

	// Ensure zone name is in the proper format (remove trailing dot for API calls)
	zone = strings.TrimSuffix(zone, ".")

	appendedRecords := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		njallaRec, err := libdnsRecordToNjalla(record, zone)
		if err != nil {
			return appendedRecords, fmt.Errorf("failed to convert record: %w", err)
		}

		// Create record in Njalla
		req := addRecordRequest{
			Domain:  zone,
			Type:    njallaRec.Type,
			Name:    njallaRec.Name,
			Content: njallaRec.Content,
			TTL:     njallaRec.TTL,
			Prio:    njallaRec.Prio,
			Weight:  njallaRec.Weight,
			Port:    njallaRec.Port,
			Target:  njallaRec.Target,
			Value:   njallaRec.Value,
		}
		var resp addRecordResponse

		if err := client.call(ctx, "add-record", req, &resp); err != nil {
			return appendedRecords, fmt.Errorf("failed to add record: %w", err)
		}

		// Convert the response back to a libdns.Record
		njallaRecord := njallaRecord(resp)

		createdRecord, err := njallaRecordToLibdns(njallaRecord)
		if err != nil {
			return appendedRecords, fmt.Errorf("failed to convert response record: %w", err)
		}

		appendedRecords = append(appendedRecords, createdRecord)
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
// This method will try to identify existing records by checking the ProviderData for record IDs,
// or by searching for matching name+type combinations. If a match is found, the record will be
// updated; otherwise, a new record will be created.
//
// The method handles timeouts properly and includes retries for transient failures.
// This method implements the libdns.RecordSetter interface.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if p.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	client := p.getClient()
	if client == nil {
		return nil, fmt.Errorf("API client not initialized")
	}

	// Check if context already has a timeout, if not add one
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	// Ensure zone name is in the proper format (remove trailing dot for API calls)
	zone = strings.TrimSuffix(zone, ".")

	// Collect records by name and type for more efficient lookups
	recordsByKey := make(map[string]libdns.Record)
	recordsWithoutIDs := make([]libdns.Record, 0)

	// First check which records already have IDs in their ProviderData
	for _, record := range records {
		id := extractRecordID(record)
		if id != "" {
			// We have an ID, so we can update directly
			recordsByKey[id] = record
		} else {
			// No ID, so we need to check if it exists
			recordsWithoutIDs = append(recordsWithoutIDs, record)
		}
	}

	// If we have records without IDs, we need to check if they exist
	existingRecordsByKey := make(map[string]libdns.Record)
	if len(recordsWithoutIDs) > 0 {
		// Use a shorter timeout for this sub-operation
		fetchCtx, fetchCancel := context.WithTimeout(ctx, 20*time.Second)
		defer fetchCancel()

		// This is unavoidable for records without IDs - we need to check if they exist
		existingRecords, err := p.GetRecords(fetchCtx, zone)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing records: %w", err)
		}

		// Build a map of existing records by name and type
		for _, rec := range existingRecords {
			key := fmt.Sprintf("%s|%s", rec.RR().Name, rec.RR().Type)
			existingRecordsByKey[key] = rec
		}
	}

	setRecords := make([]libdns.Record, 0, len(records))

	// Process records with IDs (update operations)
	for id, record := range recordsByKey {
		// Create a context with timeout for this specific operation
		opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)

		njallaRec, err := libdnsRecordToNjalla(record, zone)
		if err != nil {
			opCancel()
			return nil, fmt.Errorf("failed to convert record: %w", err)
		}

		// Update existing record
		req := editRecordRequest{
			ID:      id,
			Domain:  zone,
			Type:    njallaRec.Type,
			Name:    njallaRec.Name,
			Content: njallaRec.Content,
			TTL:     njallaRec.TTL,
			Prio:    njallaRec.Prio,
			Weight:  njallaRec.Weight,
			Port:    njallaRec.Port,
			Target:  njallaRec.Target,
			Value:   njallaRec.Value,
		}
		var resp njallaRecord

		err = client.call(opCtx, "edit-record", req, &resp)
		opCancel() // Always cancel the context when done

		if err != nil {
			return setRecords, fmt.Errorf("failed to update record %s: %w", id, err)
		}

		updatedRecord, err := njallaRecordToLibdns(resp)
		if err != nil {
			return setRecords, fmt.Errorf("failed to convert response record: %w", err)
		}

		setRecords = append(setRecords, updatedRecord)

		// Check if parent context is done after each operation
		select {
		case <-ctx.Done():
			return setRecords, ctx.Err()
		default:
			// Continue with next record
		}
	}

	// Process records without IDs - check if they exist
	for _, record := range recordsWithoutIDs {
		// Create a context with timeout for this specific operation
		opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)

		key := fmt.Sprintf("%s|%s", record.RR().Name, record.RR().Type)
		existingRecord, exists := existingRecordsByKey[key]

		njallaRec, err := libdnsRecordToNjalla(record, zone)
		if err != nil {
			opCancel()
			return nil, fmt.Errorf("failed to convert record: %w", err)
		}

		var result libdns.Record

		if exists {
			// Get ID from existing record
			id := extractRecordID(existingRecord)
			if id == "" {
				opCancel()
				return nil, fmt.Errorf("missing ID for existing record")
			}

			// Update existing record
			req := editRecordRequest{
				ID:      id,
				Domain:  zone,
				Type:    njallaRec.Type,
				Name:    njallaRec.Name,
				Content: njallaRec.Content,
				TTL:     njallaRec.TTL,
				Prio:    njallaRec.Prio,
				Weight:  njallaRec.Weight,
				Port:    njallaRec.Port,
				Target:  njallaRec.Target,
				Value:   njallaRec.Value,
			}
			var resp njallaRecord

			err = client.call(opCtx, "edit-record", req, &resp)
			opCancel()

			if err != nil {
				return setRecords, fmt.Errorf("failed to update record: %w", err)
			}

			result, err = njallaRecordToLibdns(resp)
			if err != nil {
				return setRecords, fmt.Errorf("failed to convert response record: %w", err)
			}
		} else {
			// Create new record
			req := addRecordRequest{
				Domain:  zone,
				Type:    njallaRec.Type,
				Name:    njallaRec.Name,
				Content: njallaRec.Content,
				TTL:     njallaRec.TTL,
				Prio:    njallaRec.Prio,
				Weight:  njallaRec.Weight,
				Port:    njallaRec.Port,
				Target:  njallaRec.Target,
				Value:   njallaRec.Value,
			}
			var resp addRecordResponse

			err = client.call(opCtx, "add-record", req, &resp)
			opCancel()

			if err != nil {
				return setRecords, fmt.Errorf("failed to add record: %w", err)
			}

			// Convert the response back to a libdns.Record
			njallaRecord := njallaRecord(resp)

			result, err = njallaRecordToLibdns(njallaRecord)
			if err != nil {
				return setRecords, fmt.Errorf("failed to convert response record: %w", err)
			}
		}

		setRecords = append(setRecords, result)

		// Check if parent context is done after each operation
		select {
		case <-ctx.Done():
			return setRecords, ctx.Err()
		default:
			// Continue with next record
		}
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
// This method will attempt to delete each record, using the provider-specific ID from
// the ProviderData field if available. If no ID is available, it will try to find a matching
// record in the zone by name and type.
//
// Records that couldn't be deleted (either because they don't exist or due to an error)
// will not be included in the returned slice.
//
// This method implements the libdns.RecordDeleter interface.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if p.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	client := p.getClient()
	if client == nil {
		return nil, fmt.Errorf("API client not initialized")
	}

	// Check if context already has a timeout, if not add one
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// Default timeout of 60 seconds for the entire operation
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	// Ensure zone name is in the proper format (remove trailing dot for API calls)
	zone = strings.TrimSuffix(zone, ".")

	deletedRecords := make([]libdns.Record, 0, len(records))
	recordMap := make(map[string]libdns.Record)

	// Build ID lookup for faster matching
	var idsNeeded []string
	for i, record := range records {
		id := extractRecordID(record)
		if id != "" {
			recordMap[id] = records[i]
		} else {
			// We'll need to look these up by name/type
			key := fmt.Sprintf("%s|%s", record.RR().Name, record.RR().Type)
			recordMap[key] = records[i]
			idsNeeded = append(idsNeeded, key)
		}
	}

	// Only fetch records if we need to find some by name/type
	if len(idsNeeded) > 0 {
		// Use a shorter context timeout for this sub-operation
		fetchCtx, fetchCancel := context.WithTimeout(ctx, 20*time.Second)
		defer fetchCancel()

		existingRecords, err := p.GetRecords(fetchCtx, zone)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing records: %w", err)
		}

		// Match records without IDs to existing records
		for _, existingRecord := range existingRecords {
			key := fmt.Sprintf("%s|%s", existingRecord.RR().Name, existingRecord.RR().Type)

			if _, needs := recordMap[key]; needs {
				id := extractRecordID(existingRecord)
				if id != "" {
					// Store the ID with the record for deletion
					recordMap[id] = recordMap[key]
					delete(recordMap, key)
				}
			}
		}
	}

	// Now delete records with IDs
	for id, record := range recordMap {
		// Skip non-ID keys
		if !isRecordID(id) {
			continue
		}

		// Create a context with timeout for this specific deletion
		opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)

		// Delete record
		req := removeRecordRequest{
			Domain: zone,
			ID:     id,
		}

		err := client.call(opCtx, "remove-record", req, nil)
		opCancel() // Always cancel the context when done

		if err != nil {
			return deletedRecords, fmt.Errorf("failed to delete record %s: %w", id, err)
		}

		deletedRecords = append(deletedRecords, record)

		// Check if parent context is done after each operation
		select {
		case <-ctx.Done():
			return deletedRecords, ctx.Err()
		default:
			// Continue with next record
		}
	}

	return deletedRecords, nil
}

// Helper function to extract record ID
func extractRecordID(record libdns.Record) string {
	switch r := record.(type) {
	case libdns.Address:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.CNAME:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.TXT:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.MX:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.SRV:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.ServiceBinding:
		if pd, ok := r.ProviderData.(map[string]string); ok {
			return pd["id"]
		}
	case libdns.RR:
		// libdns.RR doesn't support ProviderData
		// This is a limitation of the generic RR type
		return ""
	}
	return ""
}

// Helper function to check if a string is likely a record ID
func isRecordID(id string) bool {
	// This is a simple check to avoid using non-ID keys
	// Njalla IDs are typically not in the format "name|type"
	return !strings.Contains(id, "|")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
