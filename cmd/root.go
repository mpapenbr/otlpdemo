/*
Copyright 2023 Markus Papenbrock
*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/cmd/sample"
	"github.com/mpapenbr/otlpdemo/cmd/web"
	"github.com/mpapenbr/otlpdemo/log"
	"github.com/mpapenbr/otlpdemo/version"
)

const envPrefix = "otlpdemo"

var (
	cfgFile   string
	telemetry *config.Telemetry
)

type MyContext struct {
	context.Context
	Bla string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "otlpdemo",
	Short:   "A brief description of your application",
	Long:    ``,
	Version: version.FullVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logConfig := log.DefaultDevConfig()
		if config.LogConfig != "" {
			var err error
			logConfig, err = log.LoadConfig(config.LogConfig)
			if err != nil {
				log.Fatal("could not load log config", log.ErrorField(err))
			}
		}
		l := log.NewWithConfig(logConfig, config.LogLevel)
		cmd.SetContext(log.AddToContext(context.Background(), l))
		log.ResetDefault(l)
		ctx := MyContext{Bla: "fasel"}
		cmd.SetContext(ctx)
		// out, err := config.SetupStdOutTracing()
		if config.EnableTelemetry {
			var err error
			if telemetry, err = config.SetupTelemetry(ctx); err != nil {
				log.Error("Could not setup telemetry", log.ErrorField(err))
			}
			log.Info("Telemetry enabled")
		}
	},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	if telemetry != nil {
		telemetry.Shutdown()
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.otlpdemo.yml)")

	rootCmd.PersistentFlags().BoolVar(&config.EnableTelemetry,
		"enable-telemetry",
		false,
		"enables telemetry")
	rootCmd.PersistentFlags().StringVar(&config.TelemetryEndpoint,
		"telemetry-endpoint",
		"localhost:4317",
		"Endpoint that receives open telemetry data")
	rootCmd.PersistentFlags().StringVar(&config.LogLevel,
		"log-level",
		"info",
		"controls the log level (debug, info, warn, error, fatal)")

	// add commands here

	rootCmd.AddCommand(web.NewWebCommand())
	rootCmd.AddCommand(sample.NewSampleCommand())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".otlpdemo" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".otlpdemo")
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// we want all commands to be processed by the bindFlags function
	// even those N levels deep
	cmds := []*cobra.Command{}
	collectCommands(rootCmd, &cmds)

	bindFlags(rootCmd, viper.GetViper())
	for _, cmd := range rootCmd.Commands() {
		bindFlags(cmd, viper.GetViper())
	}
}

func collectCommands(cmd *cobra.Command, commands *[]*cobra.Command) {
	*commands = append(*commands, cmd)
	for _, subCmd := range cmd.Commands() {
		collectCommands(subCmd, commands)
	}
}

// Bind each cobra flag to its associated viper configuration
// (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their
		// equivalent keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if err := v.BindEnv(f.Name,
				fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)); err != nil {
				fmt.Fprintf(os.Stderr, "Could not bind env var %s: %v", f.Name, err)
			}
		}
		// Apply the viper config value to the flag when the flag is not set and viper
		// has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				fmt.Fprintf(os.Stderr, "Could set flag value for %s: %v", f.Name, err)
			}
		}
	})
}
