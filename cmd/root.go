package cmd

import (
	"fmt"
	"os"

	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/kiry163/image-cli/pkg/config"
	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	cfgFile     string
	verbose     bool
	quiet       bool
	recursive   bool
	noRecursive bool
	conflict    string

	appConfig config.Config
)

var rootCmd = &cobra.Command{
	Use:           "image-cli",
	Short:         "高性能图像处理命令行工具",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		showVersion, _ := cmd.Flags().GetBool("version")
		if showVersion {
			fmt.Fprintln(cmd.OutOrStdout(), Version)
			return nil
		}
		return cmd.Help()
	},
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		WriteError(os.Stderr, err)
		return apperror.ExitCode(err)
	}
	return 0
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "静默模式")
	rootCmd.PersistentFlags().BoolVar(&recursive, "recursive", true, "目录递归处理")
	rootCmd.PersistentFlags().BoolVar(&noRecursive, "no-recursive", false, "关闭递归")
	rootCmd.PersistentFlags().StringVar(&conflict, "conflict", "", "冲突策略: skip|overwrite|rename")
	rootCmd.PersistentFlags().BoolP("version", "V", false, "显示版本")
}

func initConfig(cmd *cobra.Command) error {
	if cmd.Name() == "init" {
		parent := cmd.Parent()
		if parent != nil && parent.Name() == "config" {
			return nil
		}
	}
	path, explicit := config.ConfigPath(cfgFile)
	v := config.NewViper()
	if err := config.Load(v, path, explicit); err != nil {
		return err
	}
	if cmd.Flags().Changed("conflict") {
		v.Set("base.conflict", conflict)
	}
	if cmd.Flags().Changed("recursive") {
		v.Set("base.recursive", recursive)
	}
	if noRecursive {
		v.Set("base.recursive", false)
	}
	loaded, err := config.FromViper(v)
	if err != nil {
		return err
	}
	appConfig = loaded
	return nil
}
