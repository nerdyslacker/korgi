/*
Copyright © 2020  Artyom Topchyan a.topchyan@reply.de

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
	"os"
	"path/filepath"

	"github.com/DataReply/korgi/pkg"
	"github.com/spf13/cobra"
)

func templateApp(app string, inputFilePath string, appGroupDir string, lint bool) error {

	targeAppDir := pkg.ConcatDirs(appGroupDir, app)

	err := os.MkdirAll(targeAppDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating app dir: %w", err)
	}
	if lint {
		err = templateEngine.Lint(app, inputFilePath)
		if err != nil {
			return err
		}
	}
	err = templateEngine.Template(app, inputFilePath, targeAppDir)
	if err != nil {
		return err
	}

	return nil
}

func deployAppGroup(group string, namespace string, workingDir string, filter string, lint bool, dryRun bool) error {

	namespaceDir := pkg.GetNamespaceDir(namespace)
	if _, err := os.Stat(namespaceDir); os.IsNotExist(err) {
		return fmt.Errorf("%s directory does not exist", namespaceDir)
	}

	appGroupDir := pkg.ConcatDirs(namespaceDir, group)

	targetAppGroupDir := pkg.ConcatDirs(workingDir, execTime.Format("2006-01-02/15-04:05"), namespace, group)

	err := os.MkdirAll(targetAppGroupDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating group directory: %w", err)
	}

	matches, err := filepath.Glob(appGroupDir + "/*")
	if err != nil {
		return fmt.Errorf("listing group directory: %w", err)
	}

	for _, matchedAppFile := range matches {
		appFile := filepath.Base(matchedAppFile)
		if appFile != "_app_group.yaml" {
			app := pkg.SanitzeAppName(appFile)
			if filter != "" {

				if app != filter {
					continue
				}
			}

			err = templateApp(app, matchedAppFile, targetAppGroupDir, lint)
			if err != nil {
				return fmt.Errorf("templating app: %w", err)
			}

		}

	}
	if !dryRun {
		if filter != "" {
			err = execEngine.DeployApp(group+"-"+filter, pkg.ConcatDirs(targetAppGroupDir, filter), namespace)
			if err != nil {
				return fmt.Errorf("running kapp deploy with filter: %w", err)
			}
			return nil
		}

		err = execEngine.DeployGroup(group, targetAppGroupDir, namespace)
		if err != nil {
			return fmt.Errorf("running kapp deploy: %w", err)
		}
	}

	return nil
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply resources to k8s",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		namespace, _ := cmd.Flags().GetString("namespace")

		lint, _ := cmd.Flags().GetBool("lint")

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		filter, _ := cmd.Flags().GetString("filter")

		workingDir, _ := cmd.Flags().GetString("working-dir")

		err := deployAppGroup(args[0], namespace, workingDir, filter, lint, dryRun)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().BoolP("lint", "l", false, "Lint temlate")
	applyCmd.Flags().BoolP("dry-run", "d", false, "Dry Run")
	applyCmd.Flags().StringP("namespace", "n", "", "Target namespace")
	applyCmd.MarkFlagRequired("namespace")

}
