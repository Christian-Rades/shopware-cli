package cmd

import (
	"os"
	"path/filepath"
	account_api "shopware-cli/account-api"
	"shopware-cli/extension"
	"time"

	"github.com/pkg/errors"

	termColor "github.com/fatih/color"
	"github.com/spf13/cobra"
)

var accountCompanyProducerExtensionUploadCmd = &cobra.Command{
	Use:   "upload [zip]",
	Short: "Uploads a new extension upload",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return errors.Wrap(err, "validate")
		}

		client := getAccountAPIByConfig()

		p, err := client.Producer()
		if err != nil {
			return err
		}

		zipExt, err := extension.GetExtensionByZip(path)
		if err != nil {
			return err
		}

		extName, err := zipExt.GetName()
		if err != nil {
			return err
		}

		ext, err := p.GetExtensionByName(extName)
		if err != nil {
			return err
		}

		binaries, err := p.GetExtensionBinaries(ext.Id)
		if err != nil {
			return err
		}

		zipVersion, err := zipExt.GetVersion()
		if err != nil {
			return err
		}

		var foundBinary *account_api.ExtensionBinary

		for _, binary := range binaries {
			if binary.Version == zipVersion.String() {
				foundBinary = binary
				break
			}
		}

		if foundBinary == nil {
			foundBinary, err = p.CreateExtensionBinaryFile(ext.Id, path)
			if err != nil {
				return errors.Wrap(err, "create extension binary")
			}
		} else {
			termColor.Magenta("Found a zip with version %s already. Updating it", zipVersion)
		}

		changelog, err := zipExt.GetChangelog()
		if err != nil {
			return err
		}

		avaiableVersions, err := p.GetSoftwareVersions(zipExt.GetType())
		if err != nil {
			return err
		}

		constraint, err := zipExt.GetShopwareVersionConstraint()
		if err != nil {
			return err
		}

		foundBinary.Version = zipVersion.String()
		foundBinary.Changelogs[0].Text = changelog.German
		foundBinary.Changelogs[1].Text = changelog.English
		foundBinary.CompatibleSoftwareVersions = avaiableVersions.FilterOnVersion(constraint)

		err = p.UpdateExtensionBinaryInfo(ext.Id, *foundBinary)
		if err != nil {
			return err
		}

		termColor.Green("Updated changelog. Uploading now the zip to remote")

		err = p.UpdateExtensionBinaryFile(ext.Id, foundBinary.Id, path)
		if err != nil {
			return err
		}

		termColor.Green("Submitting code review request")

		beforeReviews, err := p.GetBinaryReviewResults(ext.Id, foundBinary.Id)
		if err != nil {
			return err
		}

		err = p.TriggerCodeReview(ext.Id)
		if err != nil {
			return err
		}

		if !skipWaitingForCodereviewResult {
			termColor.Cyan("Waiting for code review result")

			time.Sleep(10 * time.Second)

			maxTries := 10
			tried := 0
			for {
				reviews, err := p.GetBinaryReviewResults(ext.Id, foundBinary.Id)
				if err != nil {
					return err
				}

				// Review has been updated
				if len(reviews) != len(beforeReviews) {
					lastReview := reviews[len(reviews)-1]

					if !lastReview.IsPending() {
						if lastReview.HasPassed() {
							if lastReview.HasWarnings() {
								termColor.Magenta("Code review has been passed but with warnings")
								termColor.Blue(lastReview.GetSummary())
							} else {
								termColor.Green("Code review has been passed without warnings")
							}

							break
						} else {
							termColor.Red("Code review has not passed")
							termColor.Red(lastReview.GetSummary())

							os.Exit(1)
						}
					}
				}

				time.Sleep(15 * time.Second)
				tried++

				if maxTries == tried {
					termColor.Green("Skipping waiting for code review result as it took too long")
				}
			}
		}

		return nil
	},
}

var skipWaitingForCodereviewResult bool

func init() {
	accountCompanyProducerExtensionCmd.AddCommand(accountCompanyProducerExtensionUploadCmd)
	accountCompanyProducerExtensionCmd.Flags().BoolVar(&skipWaitingForCodereviewResult, "skip-for-review-result", false, "Skips waiting for Code review result")
}