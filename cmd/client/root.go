package main

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/client"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/aseryang/MyTcpReverseProxy/utils/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	showVersion bool
)
var rootCmd = &cobra.Command{
	Use:                    "client",
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "client is the client of MyTcpReverseProxy(https://github.com/aseryang/MyTcpReverseproxy)",
	Long:                   "",
	Example:                "",
	ValidArgs:              nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Hidden:                 false,
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run:                    nil,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
		}
		err := runClient(cfgFile)
		if err != nil {
			return err
		}
		return nil
	},
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
	TraverseChildren:           false,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./mrpc.ini", "config file of mrp client")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of mrp client")
}
func runClient(cfgFilePath string) error {
	content, err := config.GetRenderedConfFromFile(cfgFilePath)
	if err != nil {
		return err
	}
	cfg, err := config.UnmarshalClientConfFromIni(content)
	pxyCfgs, visitorCfgs, err := config.LoadAllConfFromIni(content)
	startService(cfg, pxyCfgs, visitorCfgs)

	return nil
}
func startService(cfg config.ClientCommonConf, pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf) {
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel, cfg.LogMaxDays, false)
	svr := client.NewService(cfg, pxyCfgs, visitorCfgs)
	svr.Run()
}
