// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package apiclient contains all you need to communicate with the insanely fast HEXONET backend API.
package apiclient

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	LG "github.com/hexonet/go-sdk/logger"
	R "github.com/hexonet/go-sdk/response"
	RTM "github.com/hexonet/go-sdk/responsetemplatemanager"
	SC "github.com/hexonet/go-sdk/socketconfig"
)

// ISPAPI_CONNECTION_URL_PROXY represents the url used for the high performance connection setup
const ISPAPI_CONNECTION_URL_PROXY = "http://127.0.0.1/api/call.cgi"

// ISPAPI_CONNECTION_URL represents the url used for the default connection setup
const ISPAPI_CONNECTION_URL = "https://api.ispapi.net/api/call.cgi"

var rtm = RTM.GetInstance()

// APIClient is the entry point class for communicating with the insanely fast HEXONET backend api.
// It allows two ways of communication:
// * session based communication
// * sessionless communication
//
// A session based communication makes sense in case you use it to
// build your own frontend on top. It allows also to use 2FA
// (2 Factor Auth) by providing "otp" in the config parameter of
// the login method.
// A sessionless communication makes sense in case you do not need
// to care about the above and you have just to request some commands.
//
// Possible commands can be found at https://github.com/hexonet/hexonet-api-documentation/tree/master/API
type APIClient struct {
	socketTimeout time.Duration
	socketURL     string
	socketConfig  *SC.SocketConfig
	debugMode     bool
	curlopts      map[string]string
	ua            string
	logger        LG.ILogger
}

// NewAPIClient represents the constructor for struct APIClient.
func NewAPIClient() *APIClient {
	cl := &APIClient{
		debugMode:     false,
		socketTimeout: 300 * time.Second,
		socketURL:     ISPAPI_CONNECTION_URL,
		socketConfig:  SC.NewSocketConfig(),
		curlopts:      map[string]string{},
		ua:            "",
		logger:        nil,
	}
	cl.UseLIVESystem()
	cl.SetDefaultLogger()
	return cl
}

// SetDefaultLogger method to use the default mechanism for debug mode outputs
func (cl *APIClient) SetDefaultLogger() *APIClient {
	cl.logger = LG.NewLogger()
	return cl
}

// SetCustomLogger method to use a custom mechanism for debug mode outputs/logging
func (cl *APIClient) SetCustomLogger(logger LG.ILogger) *APIClient {
	cl.logger = logger
	return cl
}

// SetProxy method to set a proxy to use for API communication
func (cl *APIClient) SetProxy(proxy string) *APIClient {
	if len(proxy) == 0 {
		delete(cl.curlopts, "PROXY")
	} else {
		cl.curlopts["PROXY"] = proxy
	}
	return cl
}

// GetProxy method to get the configured proxy to use for API communication
func (cl *APIClient) GetProxy() (string, error) {
	val, exists := cl.curlopts["PROXY"]
	if exists {
		return val, nil
	}
	return "", errors.New("No proxy configuration available.")
}

// SetReferer method to set a value for HTTP Header `Referer` to use for API communication
func (cl *APIClient) SetReferer(referer string) *APIClient {
	if len(referer) == 0 {
		delete(cl.curlopts, "REFERER")
	} else {
		cl.curlopts["REFERER"] = referer
	}
	return cl
}

// GetReferer method to get configured HTTP Header `Referer` value
func (cl *APIClient) GetReferer() (string, error) {
	val, exists := cl.curlopts["REFERER"]
	if exists {
		return val, nil
	}
	return "", errors.New("No configuration available for HTTP Header `Referer`.")
}

// EnableDebugMode method to enable Debug Output to logger
func (cl *APIClient) EnableDebugMode() *APIClient {
	cl.debugMode = true
	return cl
}

// DisableDebugMode method to disable Debug Output to logger
func (cl *APIClient) DisableDebugMode() *APIClient {
	cl.debugMode = false
	return cl
}

// GetPOSTData method to Serialize given command for POST request
// including connection configuration data
func (cl *APIClient) GetPOSTData(cmd map[string]string, secured ...bool) string {
	data := cl.socketConfig.GetPOSTData()
	if len(secured) > 0 && secured[0] {
		re := regexp.MustCompile("s_pw=[^&]+")
		data = re.ReplaceAllString(data, "s_pw=***")
	}
	var tmp strings.Builder
	keys := []string{}
	for key := range cmd {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := cmd[key]
		tmp.WriteString(key)
		tmp.WriteString("=")
		val = strings.Replace(val, "\r", "", -1)
		val = strings.Replace(val, "\n", "", -1)
		tmp.WriteString(val)
		tmp.WriteString("\n")
	}
	str := tmp.String()
	if len(secured) > 0 && secured[0] {
		re := regexp.MustCompile("PASSWORD=[^\n]+")
		str = re.ReplaceAllString(str, "PASSWORD=***")
	}
	str = str[:len(str)-1] //remove \n at end
	return strings.Join([]string{
		data,
		url.QueryEscape("s_command"),
		"=",
		url.QueryEscape(str),
	}, "")
}

// GetSession method to get the API Session that is currently set
func (cl *APIClient) GetSession() (string, error) {
	sessid := cl.socketConfig.GetSession()
	if len(sessid) == 0 {
		return "", errors.New("Could not find an active session")
	}
	return sessid, nil
}

// GetURL method to get the API connection url that is currently set
func (cl *APIClient) GetURL() string {
	return cl.socketURL
}

// SetUserAgent method to customize user-agent header (useful for tools that use our SDK)
func (cl *APIClient) SetUserAgent(str string, rv string, modules ...[]string) *APIClient {
	mods := ""
	if len(modules) > 0 {
		for i := 0; i < len(modules[0]); i++ {
			mods += modules[0][i] + " "
		}
	}
	cl.ua = str + " (" + runtime.GOOS + "; " + runtime.GOARCH + "; rv:" + rv + ") " + mods + "go-sdk/" + cl.GetVersion() + " go/" + runtime.Version()
	return cl
}

// GetUserAgent method to return the user agent string
func (cl *APIClient) GetUserAgent() string {
	if len(cl.ua) == 0 {
		cl.ua = "GO-SDK (" + runtime.GOOS + "; " + runtime.GOARCH + "; rv:" + cl.GetVersion() + ") go/" + runtime.Version()
	}
	return cl.ua
}

// GetVersion method to get current module version
func (cl *APIClient) GetVersion() string {
	return "3.5.1"
}

// SaveSession method to apply data to a session for later reuse
// Please save/update that map into user session
func (cl *APIClient) SaveSession(sessionobj map[string]interface{}) *APIClient {
	sessionobj["socketcfg"] = map[string]string{
		"entity":  cl.socketConfig.GetSystemEntity(),
		"session": cl.socketConfig.GetSession(),
	}
	return cl
}

// ReuseSession method to reuse given configuration out of a user session
// to rebuild and reuse connection settings
func (cl *APIClient) ReuseSession(sessionobj map[string]interface{}) *APIClient {
	cfg := sessionobj["socketcfg"].(map[string]string)
	cl.socketConfig.SetSystemEntity(cfg["entity"])
	cl.SetSession(cfg["session"])
	return cl
}

// SetURL method to set another connection url to be used for API communication
func (cl *APIClient) SetURL(value string) *APIClient {
	cl.socketURL = value
	return cl
}

// SetOTP method to set one time password to be used for API communication
func (cl *APIClient) SetOTP(value string) *APIClient {
	cl.socketConfig.SetOTP(value)
	return cl
}

// SetSession method to set an API session id to be used for API communication
func (cl *APIClient) SetSession(value string) *APIClient {
	cl.socketConfig.SetSession(value)
	return cl
}

// SetRemoteIPAddress method to set an Remote IP Address to be used for API communication
func (cl *APIClient) SetRemoteIPAddress(value string) *APIClient {
	cl.socketConfig.SetRemoteAddress(value)
	return cl
}

// SetCredentials method to set Credentials to be used for API communication
func (cl *APIClient) SetCredentials(uid string, pw string) *APIClient {
	cl.socketConfig.SetLogin(uid)
	cl.socketConfig.SetPassword(pw)
	return cl
}

// SetRoleCredentials method to set Role User Credentials to be used for API communication
func (cl *APIClient) SetRoleCredentials(uid string, role string, pw string) *APIClient {
	if len(role) > 0 {
		return cl.SetCredentials(uid+"!"+role, pw)
	}
	return cl.SetCredentials(uid, pw)
}

// Login method to perform API login to start session-based communication
// 1st parameter: one time password
func (cl *APIClient) Login(params ...string) *R.Response {
	otp := ""
	if len(params) > 0 {
		otp = params[0]
	}
	cl.SetOTP(otp)
	rr := cl.Request(map[string]interface{}{"COMMAND": "StartSession"})
	if rr.IsSuccess() {
		col := rr.GetColumn("SESSION")
		if col != nil {
			cl.SetSession(col.GetData()[0])
		} else {
			cl.SetSession("")
		}
	}
	return rr
}

// LoginExtended method to perform API login to start session-based communication.
// 1st parameter: map of additional command parameters
// 2nd parameter: one time password
func (cl *APIClient) LoginExtended(params ...interface{}) *R.Response {
	otp := ""
	parameters := map[string]string{}
	if len(params) == 2 {
		otp = params[1].(string)
	}
	cl.SetOTP(otp)
	if len(params) > 0 {
		parameters = params[0].(map[string]string)
	}
	cmd := map[string]interface{}{
		"COMMAND": "StartSession",
	}
	for k, v := range parameters {
		cmd[k] = v
	}
	rr := cl.Request(cmd)
	if rr.IsSuccess() {
		col := rr.GetColumn("SESSION")
		if col != nil {
			cl.SetSession(col.GetData()[0])
		} else {
			cl.SetSession("")
		}
	}
	return rr
}

// Logout method to perform API logout to close API session in use
func (cl *APIClient) Logout() *R.Response {
	rr := cl.Request(map[string]interface{}{
		"COMMAND": "EndSession",
	})
	if rr.IsSuccess() {
		cl.SetSession("")
	}
	return rr
}

// Request method to perform API request using the given command
func (cl *APIClient) Request(cmd map[string]interface{}) *R.Response {
	// flatten nested api command bulk parameters
	newcmd := cl.flattenCommand(cmd)
	// auto convert umlaut names to punycode
	newcmd = cl.autoIDNConvert(newcmd)

	// request command to API
	cfg := map[string]string{
		"CONNECTION_URL": cl.socketURL,
	}
	data := cl.GetPOSTData(newcmd, false)
	secured := cl.GetPOSTData(newcmd, true)

	val, err := cl.GetProxy()
	client := &http.Client{
		Timeout: cl.socketTimeout,
	}
	if err == nil {
		proxyUrl, err := url.Parse(val)
		if err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		} else if cl.debugMode {
			fmt.Println("Not able to parse configured Proxy URL: " + val)
		}
	}
	req, err := http.NewRequest("POST", cfg["CONNECTION_URL"], strings.NewReader(data))
	if err != nil {
		tpl := rtm.GetTemplate("httperror").GetPlain()
		r := R.NewResponse(tpl, newcmd, cfg)
		if cl.debugMode {
			cl.logger.Log(secured, r, err.Error())
		}
		return r
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Expect", "")
	req.Header.Add("User-Agent", cl.GetUserAgent())
	val, err = cl.GetReferer()
	if err != nil {
		req.Header.Add("Referer", val)
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		tpl := rtm.GetTemplate("httperror").GetPlain()
		r := R.NewResponse(tpl, newcmd, cfg)
		if cl.debugMode {
			cl.logger.Log(secured, r, err2.Error())
		}
		return r
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tpl := rtm.GetTemplate("httperror").GetPlain()
			r := R.NewResponse(tpl, newcmd, cfg)
			if cl.debugMode {
				cl.logger.Log(secured, r, err.Error())
			}
			return r
		}
		r := R.NewResponse(string(response), newcmd, cfg)
		if cl.debugMode {
			cl.logger.Log(secured, r)
		}
		return r
	}
	tpl := rtm.GetTemplate("httperror").GetPlain()
	r := R.NewResponse(tpl, newcmd, cfg)
	if cl.debugMode {
		cl.logger.Log(secured, r)
	}
	return r
}

// RequestNextResponsePage method to request the next page of list entries for the current list query
// Useful for lists
func (cl *APIClient) RequestNextResponsePage(rr *R.Response) (*R.Response, error) {
	mycmd := map[string]interface{}{}
	for key, val := range rr.GetCommand() {
		mycmd[key] = val
	}
	if _, ok := mycmd["LAST"]; ok {
		return nil, errors.New("Parameter LAST in use. Please remove it to avoid issues in requestNextPage")
	}
	first := 0
	if v, ok := mycmd["FIRST"]; ok {
		first, _ = fmt.Sscan("%s", v)
	}
	total := rr.GetRecordsTotalCount()
	limit := rr.GetRecordsLimitation()
	first += limit
	if first < total {
		mycmd["FIRST"] = fmt.Sprintf("%d", first)
		mycmd["LIMIT"] = fmt.Sprintf("%d", limit)
		return cl.Request(mycmd), nil
	}
	return nil, errors.New("Could not find further existing pages")
}

// RequestAllResponsePages method to request all pages/entries for the given query command
// Use this method with caution as it requests all list data until done.
func (cl *APIClient) RequestAllResponsePages(cmd map[string]string) []R.Response {
	var err error
	responses := []R.Response{}
	mycmd := map[string]interface{}{}
	mycmd["FIRST"] = "0"
	for k, v := range cmd {
		mycmd[k] = v
	}
	rr := cl.Request(mycmd)
	tmp := rr
	for {
		responses = append(responses, *tmp)
		tmp, err = cl.RequestNextResponsePage(tmp)
		if err != nil {
			break
		}
	}
	return responses
}

// SetUserView method to set a data view to a given subuser
func (cl *APIClient) SetUserView(uid string) *APIClient {
	cl.socketConfig.SetUser(uid)
	return cl
}

// ResetUserView method to reset data view back from subuser to user
func (cl *APIClient) ResetUserView() *APIClient {
	cl.socketConfig.SetUser("")
	return cl
}

// UseHighPerformanceConnectionSetup to activate high performance conneciton setup
func (cl *APIClient) UseHighPerformanceConnectionSetup() *APIClient {
	cl.SetURL(ISPAPI_CONNECTION_URL_PROXY)
	return cl
}

// UseDefaultConnectionSetup to activate default conneciton setup (the default anyways)
func (cl *APIClient) UseDefaultConnectionSetup() *APIClient {
	cl.SetURL(ISPAPI_CONNECTION_URL)
	return cl
}

// UseOTESystem method to set OT&E System for API communication
func (cl *APIClient) UseOTESystem() *APIClient {
	cl.socketConfig.SetSystemEntity("1234")
	return cl
}

// UseLIVESystem method to set LIVE System for API communication
// Usage of LIVE System is active by default.
func (cl *APIClient) UseLIVESystem() *APIClient {
	cl.socketConfig.SetSystemEntity("54cd")
	return cl
}

// flattenCommand method to translate all command parameter names to uppercase
func (cl *APIClient) flattenCommand(cmd map[string]interface{}) map[string]string {
	newcmd := map[string]string{}
	for key, val := range cmd {
		newKey := strings.ToUpper(key)
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			v := val.([]string)
			for idx, str := range v {
				str = strings.Replace(str, "\r", "", -1)
				str = strings.Replace(str, "\n", "", -1)
				newcmd[newKey+strconv.Itoa(idx)] = str
			}
		} else {
			val := val.(string)
			val = strings.Replace(val, "\r", "", -1)
			val = strings.Replace(val, "\n", "", -1)
			newcmd[newKey] = val
		}
	}
	return newcmd
}

// autoIDNConvert method to translate all whitelisted parameter values to punycode, if necessary
func (cl *APIClient) autoIDNConvert(cmd map[string]string) map[string]string {
	newcmd := map[string]string{
		"COMMAND": "ConvertIDN",
	}
	// don't convert for convertidn command to avoid endless loop
	pattern := regexp.MustCompile(`(?i)^CONVERTIDN$`)
	mm := pattern.MatchString(cmd["COMMAND"])
	if mm {
		return cmd
	}
	keys := []string{}
	pattern = regexp.MustCompile(`(?i)^(DOMAIN|NAMESERVER|DNSZONE)([0-9]*)$`)
	for key := range cmd {
		mm = pattern.MatchString(key)
		if mm {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return cmd
	}
	toconvert := []string{}
	idxs := []string{}
	pattern = regexp.MustCompile(`\r|\n`)
	idnpattern := regexp.MustCompile(`(?i)[^a-z0-9. -]+`)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := pattern.ReplaceAllString(cmd[key], "")
		mm = idnpattern.MatchString(val)
		if mm {
			toconvert = append(toconvert, val)
			idxs = append(idxs, key)
		} else {
			newcmd[key] = val
		}
	}
	if len(toconvert) == 0 {
		return cmd
	}
	r := cl.Request(map[string]interface{}{
		"COMMAND": "ConvertIDN",
		"DOMAIN":  toconvert,
	})
	if !r.IsSuccess() {
		return cmd
	}
	col := r.GetColumn("ACE")
	if col != nil {
		for idx, pc := range col.GetData() {
			newcmd[idxs[idx]] = pc
		}
	}
	return newcmd
}
