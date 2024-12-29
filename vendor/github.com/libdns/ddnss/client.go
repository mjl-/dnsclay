package ddnss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/libdns/libdns"
	"golang.org/x/net/publicsuffix"
)

func (p *Provider) setRecord(ctx context.Context, zone string, record libdns.Record, clear bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// sanitize the domain, combines the zone and record names
	// the record name should typically be relative to the zone
	domain := libdns.AbsoluteName(record.Name, zone)

	params := map[string]string{"verbose": "true"}

	switch record.Type {
	case "TXT":
		params["txt"] = record.Value
	case "A":
		params["ip"] = record.Value
	case "AAAA":
		params["ipv6"] = record.Value
	default:
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	// api infos:
	// append a record: "key=$DDNSS_Token&host=$_ddnss_domain&txtm=1&txt=$txtvalue"
	// set a record: "key=$DDNSS_Token&host=$_ddnss_domain&txtm=1&txt=$txtvalue"
	// delete a record: "key=$DDNSS_Token&host=$_ddnss_domain&txtm=2"

	if clear {
		params["txtm"] = "2"
	} else {
		params["txtm"] = "1"
	}

	// make the request to ddnss to set the records according to the params
	_, err := p.doRequest(ctx, domain, params)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) doRequest(ctx context.Context, domain string, params map[string]string) ([]string, error) {
	u, _ := url.Parse("https://ddnss.de/upd.php")

	// extract the main domain
	var mainDomain string = getMainDomain(domain)

	if len(mainDomain) == 0 {
		return nil, fmt.Errorf("unable to find the main domain for: %s", domain)
	}

	// set up the query with the params we always set
	query := u.Query()
	query.Set("host", mainDomain)
	query.Set("key", p.APIToken)

	// add the remaining ones for this request
	for key, val := range params {
		query.Set(key, val)
	}

	// set the query back on the URL
	u.RawQuery = query.Encode()

	// make the request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := string(bodyBytes)

	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, errors.New(`not a valid XPath expression`)
	}

	nodes, err := htmlquery.QueryAll(doc, "//font")
	if err != nil {
		return nil, errors.New(`not a valid XPath expression`)
	}

	if len(nodes) == 0 {
		return nil, errors.New(`abort nodes`)
	}

	domains := []string{}
	for _, n := range nodes {
		if n.FirstChild == nil {
			continue
		}
		text := strings.TrimSpace(n.FirstChild.Data)
		if strings.Index(text, "Updated ") == 0 {
			domains = append(domains, domain)

			return domains, nil
		}
	}

	return nil, fmt.Errorf("DDNSS request failed, expected (OK) but got url: [%s], body: %s", u, body)
}

func getMainDomain(domain string) string {
	domain = strings.TrimSuffix(domain, ".")
	domain = strings.TrimLeft(domain, "_acme-challenge.")
	return domain
}

// There is no api for retreiving all registered domains.
// So we have to login to the ddnss.de webpage with username and password
// and extract the information from a html response
func getDomainFromWebinterface(username string, password string) ([]string, error) {
	ddnssLoginUrl := "https://ddnss.de/do.php"
	ddnssDomainListUrl := "https://ddnss.de/ua/vhosts_list.php"
	method := "POST"

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("action", "login")
	_ = writer.WriteField("username", username)
	_ = writer.WriteField("passwd", password)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	req, err := http.NewRequest(method, ddnssLoginUrl, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "golang")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	cookies := res.Cookies()
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	urlObj, _ := url.Parse(ddnssDomainListUrl)
	client.Jar.SetCookies(urlObj, cookies)

	res, err = client.Get(ddnssDomainListUrl)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, errors.New(`not a valid XPath expression.`)
	}

	nodes, err := htmlquery.QueryAll(doc, "//u")
	if err != nil {
		return nil, errors.New(`not a valid XPath expression.`)
	}

	domains := []string{}
	for _, n := range nodes {
		domain := strings.TrimSpace(n.FirstChild.Data)
		domains = append(domains, domain)
	}

	return domains, nil
}

func (p *Provider) getDomainsFromWebinterface() ([]string, error) {
	return getDomainFromWebinterface(p.Username, p.Password)
}
