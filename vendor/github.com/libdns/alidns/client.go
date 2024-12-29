package alidns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// mClient is an abstration of AliClient
type mClient struct {
	aClient *aliClient
	mutex   sync.Mutex
}

func (p *Provider) getClient() error {
	return p.getClientWithZone("")
}

func (p *Provider) getClientWithZone(zone string) error {
	cred := newCredInfo(p.AccKeyID, p.AccKeySecret, p.RegionID)
	return p.client.getAliClient(cred, zone)
}

func (p *Provider) addDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClientWithZone(rc.DName)
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	p.client.aClient.addReqBody("Action", "AddDomainRecord")
	p.client.aClient.addReqBody("DomainName", rc.DName)
	p.client.aClient.addReqBody("RR", rc.Rr)
	p.client.aClient.addReqBody("Type", rc.DTyp)
	p.client.aClient.addReqBody("Value", rc.DVal)
	p.client.aClient.addReqBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) delDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.aClient.addReqBody("Action", "DeleteDomainRecord")
	p.client.aClient.addReqBody("RecordId", rc.RecID)
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) setDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClientWithZone(rc.DName)
	p.client.aClient.addReqBody("Action", "UpdateDomainRecord")
	p.client.aClient.addReqBody("RecordId", rc.RecID)
	p.client.aClient.addReqBody("RR", rc.Rr)
	p.client.aClient.addReqBody("Type", rc.DTyp)
	p.client.aClient.addReqBody("Value", rc.DVal)
	p.client.aClient.addReqBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) getDomainRecord(ctx context.Context, recID string) (aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.aClient.addReqBody("Action", "DescribeDomainRecordInfo")
	p.client.aClient.addReqBody("RecordId", recID)
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	rec := rs.ToDomaRecord()
	if err != nil {
		return aliDomaRecord{}, err
	}
	return rec, err
}

func (p *Provider) queryDomainRecords(ctx context.Context, name string) ([]aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.aClient.addReqBody("Action", "DescribeDomainRecords")
	p.client.aClient.addReqBody("DomainName", strings.Trim(name, "."))
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return []aliDomaRecord{}, err
	}
	return rs.DRecords.Record, err
}

func (p *Provider) queryDomainRecord(ctx context.Context, rr, name string, recType ...string) (aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.aClient.addReqBody("Action", "DescribeDomainRecords")
	p.client.aClient.addReqBody("DomainName", strings.Trim(name, "."))
	p.client.aClient.addReqBody("RRKeyWord", rr)
	if len(recType) > 0 {
		p.client.aClient.addReqBody("TypeKeyWord", recType[0])
	}
	p.client.aClient.addReqBody("SearchMode", "ADVANCED")
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return aliDomaRecord{}, err
	}
	if len(rs.DRecords.Record) == 0 {
		return aliDomaRecord{}, errors.New("the Record Name of the domain not found")
	}
	return rs.DRecords.Record[0], err
}

// REVERSED:queryMainDomain rseserved for absolute names to name,zone
func (p *Provider) queryMainDomain(ctx context.Context, name string) (string, string, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.aClient.addReqBody("Action", "GetMainDomainName")
	p.client.aClient.addReqBody("InputString", strings.Trim(name, "."))
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", "", err
	}
	return rs.Rr, rs.DName, err
}

func (p *Provider) doAPIRequest(ctx context.Context, result interface{}) error {
	return p.client.doAPIRequest(ctx, "GET", result)
}

// TODO:Will complete,If we need to get Domain Info for something.
func (c *mClient) getDomainInfo(ctx context.Context, zone string) error {
	return nil
}

func (c *mClient) doAPIRequest(ctx context.Context, method string, result interface{}) error {
	req, err := c.applyReq(ctx, method, nil)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(rsp.Body)
	strBody := string(buf)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(strBody), result)
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return fmt.Errorf("get error status: HTTP %d: %+v", rsp.StatusCode, result.(*aliResult).Msg)
	}
	c.aClient = nil
	return err
}
