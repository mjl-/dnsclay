package ovh

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/libdns/libdns"
	"github.com/ovh/go-ovh/ovh"
)

type Client struct {
	ovh   *ovh.Client
	mutex sync.Mutex
}

func (p *Provider) setupClient() error {
	if p.client.ovh == nil {
		client, err := ovh.NewClient(p.Endpoint, p.ApplicationKey, p.ApplicationSecret, p.ConsumerKey)
		if err != nil {
			return err
		}

		p.client.ovh = client
	}

	return nil
}

func (p *Provider) getRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	ids, err := p.getAllRecordsIDs(ctx, zone)
	if err != nil {
		return nil, err
	}

	var records []libdns.Record
	for _, id := range ids {
		rec, err := p.getRecordByID(ctx, zone, id)
		if err != nil {
			return nil, err
		}

		records = append(records, rec)
	}

	return records, nil
}

func (p *Provider) getAllRecordsIDs(ctx context.Context, zone string) ([]int64, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return nil, err
	}

	var ids []int64
	if err := p.client.ovh.GetWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record", zone), &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func (p *Provider) getRecordByID(ctx context.Context, zone string, id int64) (libdns.Record, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return nil, err
	}

	var ovhRec ovhRecord
	if err := p.client.ovh.GetWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record/%d", zone, id), &ovhRec); err != nil {
		return nil, err
	}

	return ovhRec.libdnsRecord()
}

func (p *Provider) addRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, int64, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return nil, 0, err
	}

	ovhRec, err := toOvhRecord(record)
	if err != nil {
		return nil, 0, err
	}

	if ovhRec.FieldType == "" {
		return nil, 0, fmt.Errorf("type of record not specified")
	}

	var ovhRecAdded ovhRecord
	if err := p.client.ovh.PostWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record", zone), ovhRec, &ovhRecAdded); err != nil {
		return nil, 0, err
	}

	lrec, err := ovhRecAdded.libdnsRecord()
	return lrec, ovhRecAdded.ID, err
}

func (p *Provider) setRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	added := []libdns.Record{}
	recList := map[string][]libdns.Record{}

	var addedIds []int64
	var addErr error
	var toDeleteIds []int64

	for _, rec := range records {
		mRec := rec.RR()
		pair := fmt.Sprintf("%s/%s", mRec.Name, mRec.Type)

		if _, ok := recList[pair]; !ok {
			mRec.Data = ""
			founded, err := p.findRecords(ctx, zone, mRec)
			if err != nil {
				return nil, err
			}

			recList[pair] = []libdns.Record{}

			for idf, recf := range founded {
				exists := false
				for _, recs := range records {
					if libdnsRecordEqual(recf, recs) {
						exists = true
						break
					}
				}
				if !exists {
					toDeleteIds = append(toDeleteIds, idf)
				} else {
					recList[pair] = append(recList[pair], recf)
				}
			}
		}

		exists := false
		for _, recs := range recList[pair] {
			if libdnsRecordEqual(rec, recs) {
				exists = true
				break
			}
		}
		if !exists {
			reca, idr, err := p.addRecord(ctx, zone, rec)
			if err != nil {
				addErr = err
				break
			}
			addedIds = append(addedIds, idr)
			added = append(added, reca)
		}

	}

	if addErr != nil {
		var rollbackErrs []error
		for _, ida := range addedIds {
			if err := p.deleteRecordByID(ctx, zone, ida); err != nil {
				rollbackErrs = append(rollbackErrs, err)
			}
		}

		if len(rollbackErrs) > 0 {
			return nil, fmt.Errorf(
				"set records failed: %v; rollback failed with %d errors (possible inconsistent state on the zone): %w",
				addErr, len(rollbackErrs), errors.Join(rollbackErrs...),
			)
		}

		return nil, libdns.AtomicErr(fmt.Errorf("atomic error: %w", addErr))
	}

	var deleteErrs []error
	for _, did := range toDeleteIds {
		if err := p.deleteRecordByID(ctx, zone, did); err != nil {
			deleteErrs = append(deleteErrs, err)
		}
	}

	if len(deleteErrs) > 0 {
		return nil, fmt.Errorf(
			"set records failed during cleanup with %d errors (possible inconsistent state on the zone): %w",
			len(deleteErrs), errors.Join(deleteErrs...),
		)
	}

	return added, nil
}

func (p *Provider) deleteRecords(ctx context.Context, zone string, record libdns.Record) ([]libdns.Record, error) {
	founded, err := p.findRecords(ctx, zone, record)
	if err != nil {
		return nil, err
	}

	deleted := []libdns.Record{}
	for id, rec := range founded {
		if err := p.deleteRecordByID(ctx, zone, id); err != nil {
			return nil, err
		}
		deleted = append(deleted, rec)
	}

	return deleted, nil
}

func (p *Provider) deleteRecordByID(ctx context.Context, zone string, id int64) error {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return err
	}

	return p.client.ovh.DeleteWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record/%d", zone, id), nil)
}

func (p *Provider) findRecords(ctx context.Context, zone string, record libdns.Record) (map[int64]libdns.Record, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return nil, err
	}

	ovhMatchRec, err := toOvhRecord(record)
	if err != nil {
		return nil, err
	}

	founded := map[int64]libdns.Record{}

	var ids []int64
	if err := p.client.ovh.GetWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record?fieldType=%s&subDomain=%s", zone, ovhMatchRec.FieldType, ovhMatchRec.SubDomain), &ids); err != nil {
		return nil, err
	}

	for _, id := range ids {
		var ovhCurrentRec ovhRecord
		if err := p.client.ovh.GetWithContext(ctx, fmt.Sprintf("/domain/zone/%s/record/%d", zone, id), &ovhCurrentRec); err != nil {
			return nil, err
		}

		currentRec, err := ovhCurrentRec.libdnsRecord()
		if err != nil {
			return nil, err
		}

		if libdnsRecordMatch(record, currentRec) {
			founded[ovhCurrentRec.ID] = currentRec
		}
	}

	return founded, nil
}

func (p *Provider) refreshZone(ctx context.Context, zone string) error {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return err
	}

	if err := p.client.ovh.PostWithContext(ctx, fmt.Sprintf("/domain/zone/%s/refresh", zone), nil, nil); err != nil {
		return err
	}

	return nil
}

func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}
