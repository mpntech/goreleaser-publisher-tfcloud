/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"github.com/tradeix/goreleaser-publisher-tfcloud/pkg/provider"
	"os"
)

// privateCmd represents the private command
var privateCmd = &cobra.Command{
	Use:   "private",
	Short: "Publish a new version of the provider to the private registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		org := cmd.Flag("org").Value.String()
		if org == "" {
			org = os.Getenv("TFE_ORG")
		}
		ns := cmd.Flag("namespace").Value.String()
		if ns == "" {
			ns = os.Getenv("TFE_NAMESPACE")
		}
		keyID := cmd.Flag("key").Value.String()
		if keyID == "" {
			keyID = os.Getenv("TFE_KEYID")
		}
		token := cmd.Flag("token").Value.String()
		if token == "" {
			token = os.Getenv("TFE_TOKEN")
		}
		path := args[0]
		cfg := tfe.DefaultConfig()
		cfg.Token = token
		tfc, err := tfe.NewClient(cfg)
		if err != nil {
			fmt.Printf("failed: %w", err)
			os.Exit(1)
		}
		if err := provider.PublishPrivateProvider(ctx, tfc, org, ns, keyID, path); err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
	},
}

func init() {
	providerCmd.AddCommand(privateCmd)
	privateCmd.Flags().StringP("org", "o", "", "TFE Organization")
	privateCmd.Flags().String("namespace", "", "Registry namespace")
	privateCmd.Flags().String("key", "", "Key id used for signing")
	privateCmd.Flags().String("token", "", "Bearer token")
}
