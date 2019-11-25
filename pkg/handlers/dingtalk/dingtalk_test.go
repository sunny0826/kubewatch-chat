package dingtalk

import (
	"github.com/sunny0826/kubewatch-chart/config"
	"testing"
)

func TestDingTalk_Init(t *testing.T) {
	type fields struct {
		Token string
		Sign  string
	}
	type args struct {
		c *config.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{fields: fields{Token: "", Sign: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DingTalk{
				Token: tt.fields.Token,
				Sign:  tt.fields.Sign,
			}
			if err := d.Init(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
