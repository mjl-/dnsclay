package cloudns

import (
	"time"
)

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

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
