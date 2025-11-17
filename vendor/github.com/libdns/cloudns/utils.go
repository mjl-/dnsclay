package cloudns

import (
	"context"
	"fmt"
	"time"

	"github.com/libdns/libdns"
)

// Rounds the given TTL in seconds to the next accepted value.
// Accepted TTL values are:
//   - 60 = 1 minute
//   - 300 = 5 minutes
//   - 900 = 15 minutes
//   - 1800 = 30 minutes
//   - 3600 = 1 hour
//   - 21600 = 6 hours
//   - 43200 = 12 hours
//   - 86400 = 1 day
//   - 172800 = 2 days
//   - 259200 = 3 days
//   - 604800 = 1 week
//   - 1209600 = 2 weeks
//   - 2592000 = 1 month
//
// See https://www.cloudns.net/wiki/article/58/ for details.
func ttlRounder(ttl time.Duration) int {
	t := int(ttl.Seconds())
	for _, validTTL := range []int{60, 300, 900, 1800, 3600, 21600, 43200, 86400, 172800, 259200, 604800, 1209600} {
		if t <= validTTL {
			return validTTL
		}
	}

	return 2592000
}

// RetryWithBackoff executes the given function with exponential backoff retry logic.
// It will retry the function until it succeeds or the maximum number of retries is reached.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - operation: Function to execute
//   - maxRetries: Maximum number of retry attempts
//   - initialBackoff: Initial backoff duration
//   - maxBackoff: Maximum backoff duration
//
// Returns:
//   - error: The last error returned by the operation, or nil if it succeeded
func RetryWithBackoff(ctx context.Context, operation func() error, maxRetries int, initialBackoff, maxBackoff time.Duration) error {
	var err error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue with operation
		}

		err = operation()
		if err == nil {
			return nil // Operation succeeded
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries-1 {
			return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, err)
		}

		// Wait before retrying with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Double the backoff for next attempt, but don't exceed maxBackoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return err
}

type nameAndType struct {
	name  string
	type_ string
}

// clouDNSRecordsToMap turns a slice of raw upstream results into a map indexed
// by a the name and type of the record
func clouDNSRecordsToMap(recs []ApiDnsRecord) map[nameAndType][]ApiDnsRecord {
	ret := make(map[nameAndType][]ApiDnsRecord)
	for _, res := range recs {
		k := nameAndType{name: res.Host, type_: res.Type}
		if _, ok := ret[k]; !ok {
			ret[k] = []ApiDnsRecord{res}
		} else {
			ret[k] = append(ret[k], res)
		}
	}

	return ret
}

func libdnsRecordsToMap(recs []libdns.Record) map[nameAndType][]libdns.RR {
	ret := make(map[nameAndType][]libdns.RR)
	for _, res := range recs {
		rr := res.RR()
		k := nameAndType{name: rr.Name, type_: rr.Type}
		if _, ok := ret[k]; !ok {
			ret[k] = []libdns.RR{rr}
		} else {
			ret[k] = append(ret[k], rr)
		}
	}

	return ret
}
