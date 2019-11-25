/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/sunny0826/kubewatch-chart/config"
)

// dingtalkCmd represents the dingtalk command
var dingtalkCmd = &cobra.Command{
	Use:   "dingtalk",
	Short: "specific dingtalk configuration",
	Long:  `specific dingtalk configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			logrus.Fatal(err)
		}

		token, err := cmd.Flags().GetString("token")
		if err == nil {
			if len(token) > 0 {
				conf.Handler.Dingtalk.Token = token
			}
		} else {
			logrus.Fatal(err)
		}
		sign, err := cmd.Flags().GetString("sign")
		if err == nil {
			if len(sign) > 0 {
				conf.Handler.Dingtalk.Sign = sign
			}
		}

		if err = conf.Write(); err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	dingtalkCmd.Flags().StringP("sign", "s", "", "Specify dingtalk sign")
	dingtalkCmd.Flags().StringP("token", "t", "", "Specify dingtalk token")
}
