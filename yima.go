// Copyright 2018 Wolther47. All right reserved.
// Use of this source code is governed by a MIT-style
// License that can be found in the LICENSE.

package yima

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"

	"github.com/json-iterator/go"
)

// Yima API
const (
	YimaHost   = "api.fxhyd.cn"
	YimaScheme = "http"
	YimaPath   = "UserInterface.aspx"

	YimaCandidateHost   = "api.fxhyd.cn"
	YimaCandidateScheme = "http"
	YimaCandidatePath   = "appapi.aspx"
)

// Operator is the type of three operators in China
type Operator int

// Three operators, from 1 to 3
const (
	ChinaMobile = iota + 1
	ChinaTelecom
	ChinaUnicom
)

// Yima wraps all actions.
type Yima struct {
	Token string
}

// Login uses username & password to get the token.
func (ym *Yima) Login(username, password string) error {

	b, err := ym.get("login", map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		log.Panicf("network error when login: %v", err)
	}

	if strings.Contains(b, "success") {
		ym.Token = strings.Split(b, "|")[1]
		return nil
	}

	return fmt.Errorf("error when login, error code: %v", b)
}

// GetAccountDetail get the details of the logged in account.
func (ym *Yima) GetAccountDetail() (AccountDetail, error) {
	b, err := ym.authGET("getaccountinfo", map[string]string{
		"format": "1",
	})

	if err != nil {
		return AccountDetail{}, err

	}

	var ad AccountDetail

	if err := json.Unmarshal([]byte(b), &ad); err != nil {
		return AccountDetail{}, err
	}

	return ad, nil
}

// SearchTemplate search itemid for the SMS template.
func (ym *Yima) SearchTemplate(keyword string) ([]TemplateCandidate, error) {

	if ym.Token == "" {
		return nil, errors.New("not login")
	}

	url := url.URL{
		Host:   YimaCandidateHost,
		Scheme: YimaScheme,
		Path:   YimaCandidatePath,
	}

	query := url.Query()
	query.Set("actionid", "itemseach")
	query.Set("token", ym.Token)
	query.Set("itemname", keyword)

	url.RawQuery = query.Encode()

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c []map[string]interface{}
	var r []TemplateCandidate

	jsoniter.Get(body, "data", "list").ToVal(&c)

	for _, i := range c {
		n := TemplateCandidate{
			ID:       int(i["ID"].(float64)),
			ItemName: i["ItemName"].(string),
			Price:    i["Price"].(float64),
			Regex:    i["Regex"].(string),
		}
		r = append(r, n)
	}

	return r, nil

}

// GetNumber requests the available number.
func (ym *Yima) GetNumber(itemID int, option *MobileOption) (string, error) {

	params := map[string]string{}

	params["itemid"] = strconv.Itoa(itemID)

	s := structs.Map(*option)

	for k, v := range s {
		switch v.(type) {
		case string:
			params[k] = v.(string)
		case int:
			params[k] = strconv.Itoa(v.(int))
		case Operator:
			params[k] = string(v.(Operator))
		default:
			return "", fmt.Errorf("options type error: %v", reflect.TypeOf(v))
		}
	}

	b, err := ym.authGET("getmobile", params)
	if err != nil {
		return "", err
	}

	return b, nil
}

// GetSMSMessage pulls the SMS authentication code.
func (ym *Yima) GetSMSMessage(phoneNo string, itemID int, addignore bool) (string, error) {

	params := map[string]string{
		"itemid": strconv.Itoa(itemID),
		"mobile": phoneNo,
	}

	if addignore {
		params["release"] = "1"
	}

	b, err := ym.authGET("getsms", params)

	if err != nil {
		return "", err
	}

	return b, nil
}

// SendSMSCode sends a SMS message.
func (ym *Yima) SendSMSCode(phoneNo string, itemID int, text string) error {

	params := map[string]string{
		"itemid": strconv.Itoa(itemID),
		"mobile": phoneNo,
		"sms":    text,
	}

	_, err := ym.authGET("sendsms", params)
	if err != nil {
		return err
	}

	return nil
}

// GetSentSMSStatus requests the status of a sent SMS message.
func (ym *Yima) GetSentSMSStatus(phoneNo string, itemID int) error {

	_, err := ym.authGET("getsendsmsstate", map[string]string{
		"itemid": strconv.Itoa(itemID),
		"mobile": phoneNo,
	})

	if err != nil {
		return err
	}

	return nil
}

// ReleaseNumber release a certain number.
func (ym *Yima) ReleaseNumber(phoneNo string, itemID int) error {
	_, err := ym.authGET("release", map[string]string{
		"itemid": strconv.Itoa(itemID),
		"mobile": phoneNo,
	})

	if err != nil {
		return err
	}
	return nil
}

// BlockNumber blocks certain number.
func (ym *Yima) BlockNumber(phoneNo string, itemID int) error {
	_, err := ym.authGET("addignore", map[string]string{
		"itemid": strconv.Itoa(itemID),
		"mobile": phoneNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (ym *Yima) get(action string, params map[string]string) (string, error) {

	url := url.URL{
		Host:   YimaHost,
		Scheme: YimaScheme,
		Path:   YimaPath,
	}

	query := url.Query()
	for k, v := range params {
		query.Add(k, v)
	}

	query.Add("action", action)

	url.RawQuery = query.Encode()

	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}

func (ym *Yima) authGET(action string, params map[string]string) (string, error) {
	if ym.Token == "" {
		return "", errors.New("not login")
	}

	params["token"] = ym.Token

	s, err := ym.get(action, params)

	if strings.Contains(s, "success") {
		return strings.Split(s, "|")[1], nil
	}

	return s, err
}

// AccountDetail represents the account details.
type AccountDetail struct {
	Name     string `json:"UserName"`
	Level    int    `json:"UserLevel"`
	Balance  float32
	Frozen   float32
	Discount float32
	MaxHold  int
	Status   int
}

// TemplateCandidate is
type TemplateCandidate struct {
	ID       int
	ItemName string
	Price    float64
	Regex    string
}

// MobileOption sets the limitation of mobile numbers.
type MobileOption struct {
	ISP       Operator
	Province  string
	City      string
	Mobile    string
	ExcludeNo string
}


func Between(str, starting, ending string) string {
	s := strings.Index(str, starting)
	if s < 0 {
		return ""
	}
	s += len(starting)
	e := strings.Index(str[s:], ending)
	if e < 0 {
		return ""
	}
	return str[s : s+e]
}
