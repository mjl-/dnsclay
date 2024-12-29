package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mjl-/bstore"
)

var refreshReschedule = make(chan struct{}, 1)

func refreshKick() {
	select {
	case refreshReschedule <- struct{}{}:
	default:
	}
}

// Refresher sleeps most of the time, until it is time to poll for a new SOA record
// with a remote zone, or for a schedule full sync of the records.
func refresher() {
	log := slog.Default()

	sync := time.NewTimer(0)
	soaCheck := time.NewTimer(0)

	// Reschedule based on the first next time to check for a new SOA record or do a full sync.
	reschedule := func() {
		err := database.Read(shutdownCtx, func(tx *bstore.Tx) error {
			if z, err := bstore.QueryTx[Zone](tx).SortAsc("NextSync").Limit(1).Get(); err == nil {
				sync.Reset(time.Until(z.NextSync))
				log.Debug("next automatic sync", "wait", time.Until(z.NextSync))
			} else if err != bstore.ErrAbsent {
				return fmt.Errorf("query zones for next automatic sync: %w", err)
			}

			if z, err := bstore.QueryTx[Zone](tx).SortAsc("NextRefresh").Limit(1).Get(); err == nil {
				soaCheck.Reset(time.Until(z.NextRefresh))
				log.Debug("next automatic refresh", "wait", time.Until(z.NextRefresh))
			} else if err != bstore.ErrAbsent {
				return fmt.Errorf("query zones for next automatic refresh: %w", err)
			}

			return nil

		})
		logCheck(log, err, "rescheduling")
	}

	reschedule()

	for {
		select {
		case <-sync.C:
			// Should be time to do a sync. Multiple zones may be ready. Fetch them and update
			// their next times to sync.
			var zones []Zone
			err := database.Write(shutdownCtx, func(tx *bstore.Tx) error {
				now := time.Now()
				var err error
				zones, err = bstore.QueryTx[Zone](tx).FilterLessEqual("NextSync", now).List()
				if err != nil {
					return fmt.Errorf("listing zones to sync: %w", err)
				}

				for _, z := range zones {
					z.NextSync = now.Add(max(z.SyncInterval, time.Minute))
					if err := tx.Update(&z); err != nil {
						return fmt.Errorf("setting next automatic sync for zone: %w", err)
					}
				}

				return nil
			})
			logCheck(log, err, "fetching zones for automatic sync")

			// With new next times, we can reschedule.
			reschedule()

			// Sync each of the zones we found.
			go func() {
				for i, z := range zones {
					// Prevent thundering herd at restart.
					if i > 0 {
						time.Sleep(2 * time.Second)
					}
					go func() {
						defer recoverPanic(log, "automatic zone refresh")
						err := refreshZoneSync(log, z)
						logCheck(log, err, "automatic zone refresh", "zone", z)
					}()
				}
			}()

		case <-soaCheck.C:
			// Gather zones to check, while updating them with a new "next" time to check.
			var zones []Zone
			err := database.Write(shutdownCtx, func(tx *bstore.Tx) error {
				now := time.Now()
				var err error
				zones, err = bstore.QueryTx[Zone](tx).FilterLessEqual("NextRefresh", now).List()
				if err != nil {
					return fmt.Errorf("listing zones to soa-check: %w", err)
				}

				// If the last record change for a zone was recent, we check a bit more often for a
				// while. More changes may be coming. DNS records are often stable for a long time,
				// until an admin starts to do some work, making multiple changes.
				for _, z := range zones {
					ival := max(z.RefreshInterval, time.Minute)
					var t time.Time
					if z.LastRecordChange != nil && time.Since(*z.LastRecordChange) < ival {
						// Iterate through the times, and pick the first that is after current time.
						t = *z.LastRecordChange
						for i := 0; i < 5 && !t.After(now); i++ {
							t = t.Add(ival / 50)
						}
						for i := 0; i < 9 && !t.After(now); i++ {
							t = t.Add(ival / 10)
						}
					} else {
						t = now.Add(ival)
					}
					z.NextRefresh = t
					if err := tx.Update(&z); err != nil {
						return fmt.Errorf("setting next automatic refresh for zone: %w", err)
					}
				}

				return nil
			})
			logCheck(log, err, "fetching zones for automatic refresh")

			// Reschedule with new next times.
			reschedule()

			go func() {
				for i, z := range zones {
					// Prevent thundering herd at restart.
					if i > 0 {
						time.Sleep(2 * time.Second)
					}
					go func() {
						defer recoverPanic(log, "automatic zone refresh")
						err := refreshZoneSOACheck(log, z)
						logCheck(log, err, "automatic zone refresh", "zone", z)
					}()
				}
			}()

		case <-refreshReschedule:
			// Something changed about a zone, e.g. zone added/removed, sync done.
			reschedule()
		}
	}
}

// refreshZoneSync fetches new records from the provider and updates local records.
func refreshZoneSync(log *slog.Logger, z Zone) error {
	pc := ProviderConfig{Name: z.ProviderConfigName}
	if err := database.Get(shutdownCtx, &pc); err != nil {
		return fmt.Errorf("get provider config: %v", err)
	}
	provider, err := providerForConfig(pc.ProviderName, pc.ProviderConfigJSON)
	if err != nil {
		return fmt.Errorf("making provider for zone: %w", err)
	}

	ctx, cancel := context.WithTimeout(shutdownCtx, 30*time.Second)
	defer cancel()

	unlock := lockZone(z.Name)
	defer unlock()

	// Get latest.
	latest, err := getRecords(ctx, log, provider, z.Name)
	if err != nil {
		return fmt.Errorf("get latest records: %w", err)
	}

	var notify bool
	defer possiblyZoneNotify(log, z.Name, &notify)

	return database.Write(ctx, func(tx *bstore.Tx) error {
		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		if err != nil {
			return fmt.Errorf("updating state with latest records: %w", err)
		}
		return nil
	})
}

// refreshZoneSOACheck fetches the zone SOA record directly from the public
// authoritative servers for a zone. If it has a different serial from what we
// think the remote zone has, it starts a full record sync.
func refreshZoneSOACheck(log *slog.Logger, z Zone) error {
	ctx, cancel := context.WithTimeout(shutdownCtx, 30*time.Second)
	defer cancel()
	lsoa, err := getSOA(ctx, log, z.Name)
	if err != nil {
		return fmt.Errorf("get latest soa from authoritative name servers: %v", err)
	}

	logCheck(log, err, "reading zone soa for refresh")
	if z.SerialRemote == Serial(lsoa.Serial) {
		log.Debug("zone refreshindicates still up to date", "zone", z.Name)
		return nil
	}

	log.Debug("refresh indicates zone has changed")
	return refreshZoneSync(log, z)
}
