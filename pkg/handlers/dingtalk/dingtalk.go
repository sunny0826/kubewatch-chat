/*
Copyright 2016 Skippbox, Ltd.

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

package dingtalk

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sunny0826/dingtalk"

	"github.com/sunny0826/kubewatch-chart/config"
	kbEvent "github.com/sunny0826/kubewatch-chart/pkg/event"
)

var dingColors = map[string]string{
	"Normal":  "#67C23A",
	"Warning": "#E6A23C",
	"Danger":  "#F56C6C",
}

var cnAction = map[string]string{
	"created": "新建",
	"deleted": "删除",
	"updated": "更新",
}

var dingErrMsg = `
%s

You need to set both dingtalk token and sign(Optional) for dingtalk notify,
using "--token/-t" and "--sign/-s", or using environment variables:

export KW_DINGTALK_TOKEN=dingtalk_token
export KW_DINGTALK_SIGN=dingtalk_sign

Command line flags will override environment variables

`

// DingTalk handler implements handler.Handler interface,
type DingTalk struct {
	Token string
	Sign  string
}

type DingContent struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Action  string `json:"action"`
	Kind    string `json:"kind"`
}

// Init prepares slack configuration
func (d *DingTalk) Init(c *config.Config) error {
	token := c.Handler.Dingtalk.Token
	sign := c.Handler.Dingtalk.Sign

	if token == "" {
		token = os.Getenv("KW_DINGTALK_TOKEN")
	}

	if sign == "" {
		sign = os.Getenv("KW_DINGTALK_SIGN")
	}

	d.Token = token
	d.Sign = sign

	return checkMissingSlackVars(d)
}

// ObjectCreated calls notifySlack on event creation
func (d *DingTalk) ObjectCreated(obj interface{}) {
	notifySlack(d, obj, "created")
}

// ObjectDeleted calls notifySlack on event creation
func (d *DingTalk) ObjectDeleted(obj interface{}) {
	notifySlack(d, obj, "deleted")
}

// ObjectUpdated calls notifySlack on event creation
func (d *DingTalk) ObjectUpdated(oldObj, newObj interface{}) {
	notifySlack(d, newObj, "updated")
}

// TestHandler tests the handler configurarion by sending test messages.
func (d *DingTalk) TestHandler() {
	webHook := dingtalk.NewWebHook(d.Token, d.Sign)

	content := DingContent{
		Title:   "kubewatch",
		Message: "Testing Handler Configuration. This is a Test message.",
		Kind:    "test",
		Action:  "created",
	}

	color := dingColors["Normal"]

	tpl := content.sendDingContent(color)

	err := webHook.SendMarkdownMsg(content.Title, tpl, false, "")
	if nil != err {
		log.Printf("%s\n", err)
		return
	}

	lasttime := time.Now().Format("2006-01-02 15:04:05")

	log.Printf("Message successfully sent to dingtalk at %s", lasttime)
}

func notifySlack(d *DingTalk, obj interface{}, action string) {
	webHook := dingtalk.NewWebHook(d.Token, d.Sign)
	e := kbEvent.New(obj, action)

	content := DingContent{
		Title:   "kubewatch",
		Message: e.DingMessage(),
		Action:  e.Reason,
		Kind:    e.Kind,
	}

	color := dingColors[e.Status]

	tpl := content.sendDingContent(color)

	err := webHook.SendMarkdownMsg(content.Title, tpl, false, "")
	if nil != err {
		log.Printf("%s\n", err)
		return
	}

	lasttime := time.Now().Format("2006-01-02 15:04:05")

	log.Printf("Message successfully sent to dingtalk at %s", lasttime)
}

func checkMissingSlackVars(d *DingTalk) error {
	if d.Token == "" {
		return fmt.Errorf(dingErrMsg, "Missing dingtalk token")
	}

	return nil
}

func (s *DingContent) sendDingContent(color string) string {
	var tpl string

	// title
	title := fmt.Sprintf("<font color=%s>%s-%s</font>", color, s.Kind, cnAction[s.Action])
	tpl = fmt.Sprintf("# %s \n", title)

	// message
	message := fmt.Sprintf("%s \n", s.Message)
	tpl += message

	return tpl
}
