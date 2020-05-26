package witai

// api doc https://wit.ai/docs/http/20170307

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ConfVal struct {
	Name string  `json:"value"`
	Conf float64 `json:"confidence"`
}

type Entities struct {
	Intent []*ConfVal `json:"intent"`
}

type Response struct {
	MsgId    string `json:"msg_id"`
	Query    string `json:"_text"`
	Entities Entities
}

type Client struct {
	token     string
	Timeout   time.Duration
	Threshold float64
}

func New(token string) *Client {

	return &Client{
		token:     token,
		Timeout:   time.Second * 10,
		Threshold: 0.5,
	}

}

func (c *Client) prepareMessage(txt string) string {

	builder := strings.Builder{}
	i := 0

	for _, rune := range txt {
		builder.WriteRune(rune)
		i++

		if i >= 260 {
			builder.WriteString("...")
			break
		}
	}
	return builder.String()
}

func (c *Client) Message(txt string) (lst []string, e error) {

	defer func() {

		if r := recover(); r != nil {
			lst = nil
			e = errors.New(fmt.Sprint(r))
		}

	}()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	ua := &http.Client{
		Timeout:   c.Timeout,
		Transport: tr,
	}

	tr.DisableKeepAlives = true

	req, err := http.NewRequest("GET", "https://api.wit.ai/message?v=20170307&q="+url.QueryEscape(c.prepareMessage(txt)), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, e := ua.Do(req)
	if e != nil {
		return nil, e
	}

	if resp.StatusCode != 200 {
		cont, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("wit.ai error: " + string(cont))
	}

	body, e1 := ioutil.ReadAll(resp.Body)
	if e1 != nil {
		return nil, errors.New("wit.ai error: " + e1.Error())
	}

	var r Response

	if e := json.Unmarshal(body, &r); e != nil {
		return nil, errors.New("wit.ai error: " + e.Error())
	}

	lst = make([]string, 0, len(r.Entities.Intent))

	for _, v := range r.Entities.Intent {

		if v.Conf >= c.Threshold {
			lst = append(lst, v.Name)
		}
	}

	return lst, nil
}
