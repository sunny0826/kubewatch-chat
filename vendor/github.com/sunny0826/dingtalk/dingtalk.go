package dingtalk

/*
Copyright © 2019 Guo Xudong

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

// LinkMsg `link message struct`
type LinkMsg struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageUrl"`
	PicURL     string `json:"picUrl"`
}

// ActionCard `action card message struct`
type ActionCard struct {
	Text           string `json:"text"`
	Title          string `json:"title"`
	SingleTitle    string `json:"singleTitle"`
	SingleURL      string `json:"singleUrl"`
	BtnOrientation string `json:"btnOrientation"`
	HideAvatar     string `json:"hideAvatar"` //  robot message avatar
	Buttons        []struct {
		Title     string `json:"title"`
		ActionURL string `json:"actionUrl"`
	} `json:"btns"`
}

// PayLoad payload
type PayLoad struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	Link struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		PicURL     string `json:"picURL"`
		MessageURL string `json:"messageUrl"`
	} `json:"link"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	ActionCard ActionCard `json:"actionCard"`
	FeedCard   struct {
		Links []LinkMsg `json:"links"`
	} `json:"feedCard"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	} `json:"at"`
}

// WebHook `web hook base config`
type WebHook struct {
	AccessToken string `json:"accessToken"`
	APIURL      string `json:"apiUrl"`
	Sign        string `json:"sign"`
	Timestamp   int64  `json:"timestamp"`
}

// Response `DingTalk web hook response struct`
type Response struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
}

// NewWebHook `new a WebHook`
func NewWebHook(accessToken string, secret string) *WebHook {
	baseAPI := "https://oapi.dingtalk.com/robot/send?access_token="
	timestamp := time.Now().UnixNano() / 1e6
	var sign string
	if secret != "" {
		sign = signWebHook(timestamp, secret)
	}
	return &WebHook{AccessToken: accessToken, APIURL: baseAPI, Timestamp: timestamp, Sign: sign}
}

// reset api URL
func (w *WebHook) resetAPIURL() {
	w.APIURL = "https://oapi.dingtalk.com/robot/send?access_token="
}

var regStr = `^1([38][0-9]|14[57]|5[^4])\d{8}$`
var regPattern = regexp.MustCompile(regStr)

// sign dingtalk webhook
func signWebHook(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}

//  real send request to api
func (w *WebHook) sendPayload(payload *PayLoad) error {
	//apiURL := w.APIURL + w.AccessToken
	apiURL := fmt.Sprintf("%v%v&timestamp=%v&sign=%v", w.APIURL, w.AccessToken, w.Timestamp, w.Sign)

	//  get config
	bs, _ := json.Marshal(payload)
	//  request api
	// log.Println(string(bs))

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(bs))

	if nil != err {
		return errors.New("api request error: " + err.Error())
	}

	//  read response body
	body, _ := ioutil.ReadAll(resp.Body)

	//  api unusual
	if 200 != resp.StatusCode {
		return fmt.Errorf("api response error: %d", resp.StatusCode)
	}

	var result Response
	//  json decode
	err = json.Unmarshal(body, &result)
	if nil != err {
		return errors.New("response struct error: response is not a json anymore, " + err.Error())
	}

	if 0 != result.ErrorCode {
		return fmt.Errorf("api custom error: {code: %d, msg: %s}", result.ErrorCode, result.ErrorMessage)
	}

	return nil
}

// SendTextMsg `send a text message`
func (w *WebHook) SendTextMsg(content string, isAtAll bool, mobiles ...string) error {
	//  send request
	return w.sendPayload(&PayLoad{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: content,
		},
		At: struct {
			AtMobiles []string `json:"atMobiles"`
			IsAtAll   bool     `json:"isAtAll"`
		}{
			AtMobiles: mobiles,
			IsAtAll:   isAtAll,
		},
	})
}

// SendLinkMsg `send a link message`
func (w *WebHook) SendLinkMsg(title, content, picURL, msgURL string) error {
	return w.sendPayload(&PayLoad{
		MsgType: "link",
		Link: struct {
			Title      string `json:"title"`
			Text       string `json:"text"`
			PicURL     string `json:"picURL"`
			MessageURL string `json:"messageUrl"`
		}{
			Title:      title,
			Text:       content,
			PicURL:     picURL,
			MessageURL: msgURL,
		},
	})
}

// SendMarkdownMsg `send a markdown msg`
func (w *WebHook) SendMarkdownMsg(title, content string, isAtAll bool, mobiles ...string) error {
	firstLine := false
	for _, mobile := range mobiles {
		if regPattern.MatchString(mobile) {
			if false == firstLine {
				content += "#####"
			}
			content += " @" + mobile
			firstLine = true
		}
	}
	//  send request
	return w.sendPayload(&PayLoad{
		MsgType: "markdown",
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{
			Title: title,
			Text:  content,
		},
		At: struct {
			AtMobiles []string `json:"atMobiles"`
			IsAtAll   bool     `json:"isAtAll"`
		}{
			AtMobiles: mobiles,
			IsAtAll:   isAtAll,
		},
	})
}

// SendActionCardMsg `send single action card message`
func (w *WebHook) SendActionCardMsg(title, content string, linkTitles, linkUrls []string, hideAvatar, btnOrientation bool) error {
	//  validation is empty
	if 0 == len(linkTitles) || 0 == len(linkUrls) {
		return errors.New("links or titles is empty！")
	}
	//  validation is equal
	if len(linkUrls) != len(linkTitles) {
		return errors.New("links length and titles length is not equal！")
	}
	//  hide robot avatar
	var strHideAvatar = "0"
	if hideAvatar {
		strHideAvatar = "1"
	}
	//  button sort
	var strBtnOrientation = "0"
	if btnOrientation {
		strBtnOrientation = "1"
	}
	//  button struct
	var buttons []struct {
		Title     string `json:"title"`
		ActionURL string `json:"actionUrl"`
	}
	//  inject to button
	for i := 0; i < len(linkTitles); i++ {
		buttons = append(buttons, struct {
			Title     string `json:"title"`
			ActionURL string `json:"actionUrl"`
		}{
			Title:     linkTitles[i],
			ActionURL: linkUrls[i],
		})
	}
	//  send request
	return w.sendPayload(&PayLoad{
		MsgType: "actionCard",
		ActionCard: ActionCard{
			Title:          title,
			Text:           content,
			HideAvatar:     strHideAvatar,
			BtnOrientation: strBtnOrientation,
			Buttons:        buttons,
		},
	})
}

// SendLinkCardMsg `send link card message`
func (w *WebHook) SendLinkCardMsg(messages []LinkMsg) error {
	return w.sendPayload(&PayLoad{
		MsgType: "feedCard",
		FeedCard: struct {
			Links []LinkMsg `json:"links"`
		}{
			Links: messages,
		},
	})
}
