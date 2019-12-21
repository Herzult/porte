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
	"net/url"

	"github.com/herzult/porte/internal/graph/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/herzult/porte/internal/graph/debug"
	"github.com/herzult/porte/internal/graph/playground"

	"github.com/spf13/viper"

	"github.com/herzult/porte/internal/graph"
	"github.com/herzult/porte/internal/graph/proxy"

	"github.com/spf13/cobra"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Starts a GraphQL proxy",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		graphURL, err := url.Parse(viper.GetString("proxy.graph-url"))
		if err != nil {
			panic(err)
		}
		g, err := graph.NewGraph(&graph.GraphConfig{
			ServiceURL: graphURL,
		})
		if err != nil {
			panic(err)
		}

		plugs := make([]*proxy.Plugin, 0)
		if viper.GetBool("proxy.debug") {
			plug, _ := debug.NewProxyPlugin()
			plugs = append(plugs, plug)
		}
		if viper.GetBool("proxy.playground") {
			plug, _ := playground.NewProxyPlugin()
			plugs = append(plugs, plug)
		}
		if viper.GetBool("proxy.prometheus") {
			plug, _ := prometheus.NewProxyPlugin(prometheus.ProxyPluginConfig{
				Namespace: "porte",
				Subsystem: "proxy",
			})
			plugs = append(plugs, plug)
			http.Handle("/metrics", promhttp.Handler())
		}

		p, err := proxy.New(&proxy.Config{
			Graph:   g,
			Plugins: plugs,
		})
		if err != nil {
			panic(err)
		}

		http.Handle(viper.GetString("proxy.path"), p)

		address := fmt.Sprint(":", viper.GetString("proxy.port"))
		cmd.Printf("Listening and serving HTTP on %s\n", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			cmd.PrintErr("failed to run proxy: ", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().String("port", "8080", "Port to run proxy on")
	proxyCmd.Flags().String("path", "/graphql", "Path to handle GraphQL requests on")
	proxyCmd.Flags().String("graph-url", "", "URL of the GraphQL service")
	proxyCmd.Flags().Bool("playground", false, "Enable the GraphQL playground")
	proxyCmd.Flags().Bool("debug", false, "Enable debug mode")
	proxyCmd.Flags().Bool("prometheus", false, "Enable prometheus on /metrics")

	viper.BindPFlag("proxy.port", proxyCmd.Flags().Lookup("port"))
	viper.BindPFlag("proxy.path", proxyCmd.Flags().Lookup("path"))
	viper.BindPFlag("proxy.graph-url", proxyCmd.Flags().Lookup("graph-url"))
	viper.BindPFlag("proxy.playground", proxyCmd.Flags().Lookup("playground"))
	viper.BindPFlag("proxy.debug", proxyCmd.Flags().Lookup("debug"))
	viper.BindPFlag("proxy.prometheus", proxyCmd.Flags().Lookup("prometheus"))
}
