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
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/handler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fakegraphCmd represents the fakegraph command
var fakegraphCmd = &cobra.Command{
	Use:   "fakegraph",
	Short: "Runs a sample graph.",
	Run: func(cmd *cobra.Command, args []string) {
		http.Handle("/graphql", handler.New(&handler.Config{
			Schema: &testutil.StarWarsSchema,
		}))

		address := fmt.Sprint(":", viper.GetString("fakegraph.port"))
		cmd.Printf("Listening and serving HTTP on %s\n", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			cmd.PrintErr("failed to run fake graph: ", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(fakegraphCmd)
	fakegraphCmd.Flags().String("port", "8888", "Port to run the fake graph on")
	viper.BindPFlag("fakegraph.port", fakegraphCmd.Flags().Lookup("port"))
}
