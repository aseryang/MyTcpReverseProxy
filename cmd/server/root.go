package main

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/server"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/aseryang/MyTcpReverseProxy/utils/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	showVersion bool
)
var rootCmd = &cobra.Command{
	Use:                    "server",
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "server is the server of MyTcpReverseProxy(https://github.com/aseryang/MyTcpReverseproxy)",
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
		err := runServer(cfgFile)
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./mrps.ini", "config of mrp server")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of mrp server")

}
func runServer(cfgFilePath string) error {
	content, err := config.GetRenderedConfFromFile(cfgFilePath)
	if err != nil {
		return err
	}
	cfg, err := config.UnmarshalServerConfFromIni(content)
	if err != nil {
		return err
	}
	startService(cfg)
	return nil
}
func startService(cfg config.ServerCommonConf) {
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel, cfg.LogMaxDays, false)
	err, svr := server.NewService(cfg)
	if err != nil {
		return
	}else {
		svr.Run()
	}
}
