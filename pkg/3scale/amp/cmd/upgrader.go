package cmd

import (
	"fmt"
	"os"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/cmd/upgrader"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade your 3scale installation",
	Long:  "Upgrade your 3scale installation",
	Run:   upgradeCommandEntrypoint,
}

func upgradeCommandEntrypoint(cmd *cobra.Command, args []string) {
	upgradeScheme := scheme.Scheme
	if err := upgrader.RegisterOpenShiftAPIGroups(upgradeScheme); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprint(err))
		os.Exit(1)
	}

	configuration := config.GetConfigOrDie()
	cl, err := client.New(configuration, client.Options{Scheme: upgradeScheme})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create client: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.CheckCurrentInstallation(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check current 3scale installation: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.MigrateSystemSMTPData(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate System SMTP data: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.UpgradeSystemPreHook(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate System pre hook config: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.UpgradeSystemAppContainerEnvs(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate System containers env: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.UpgradeSystemSidekiqEnvs(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate System sidekiq env: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.MigrateS3Deployment(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate S3 deployment: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.PatchAMPRelese(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to patch amp release: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.UpgradeImages(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to upgrade images: %v\n", err)
		os.Exit(1)
	}

	err = upgrader.DeleteSMTPConfigMap(cl, upgradeNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete smtp config map: %v\n", err)
		os.Exit(1)
	}

	if !upgrader.VerifyUpgrade(cl, upgradeNamespace) {
		os.Exit(1)
	}

	fmt.Println("3scale successfully upgraded to release 2.8")
}

var upgradeNamespace string

func init() {
	upgradeCmd.Flags().StringVarP(&upgradeNamespace, "namespace", "n", "", "Cluster namespace (required)")
	upgradeCmd.MarkFlagRequired("namespace")
	rootCmd.AddCommand(upgradeCmd)

}
