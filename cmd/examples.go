package cmd

import (
	"fmt"
	"github.com/co-pilot-cli/co-pilot/pkg/file"
	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "examples found in cloud-config",
	Long:  `examples found in cloud-config`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := EnableDebug(cmd); err != nil {
			log.Fatalln(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// sync cloud config
		if err := activeCloudConfig.Refresh(activeLocalConfig); err != nil {
			log.Fatalln(err)
		}

		examples, err := activeCloudConfig.Examples()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Available examples are:")
		for _, example := range examples {
			fmt.Printf("\t* %s\n", example)
		}
	},
}

var examplesInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install example from cloud-config",
	Long:  `install example from cloud-config`,
	Run: func(cmd *cobra.Command, args []string) {
		exampleName, _ := cmd.Flags().GetString("name")

		if exampleName == "" {
			log.Fatalln("please enter example --name")
		}

		// sync cloud config
		if err := activeCloudConfig.Refresh(activeLocalConfig); err != nil {
			log.Fatalln(err)
		}

		examples, err := activeCloudConfig.Examples()
		if err != nil {
			log.Fatalln(err)
		}

		for _, example := range examples {
			if example == exampleName {
				path := file.Path("%s/examples/%s/co-pilot.json", activeCloudConfig.Implementation().Dir(), example)
				cmd.Flags().String("config-file", path, "Optional config file")
				generateCmd.Run(cmd, args)
				return
			}
		}

		log.Fatalf("could not find %s in examples", exampleName)
	},
}

func init() {
	RootCmd.AddCommand(examplesCmd)

	examplesCmd.PersistentFlags().StringVar(&ctx.TargetDirectory, "target", ".", "Optional target directory")
	examplesCmd.PersistentFlags().BoolVar(&ctx.DryRun, "dry-run", false, "dry run does not write to pom.xml")

	examplesCmd.AddCommand(examplesInstallCmd)
	examplesInstallCmd.Flags().StringP("name", "n", "", "Example name")
	examplesInstallCmd.Flags().String("group-id", "", "Overrides groupId from config file")
	examplesInstallCmd.Flags().String("artifact-id", "", "Overrides artifactId from config file")
}