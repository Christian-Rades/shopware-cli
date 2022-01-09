package cmd

import (
	"errors"
	"fmt"
	termColor "github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	accountApi "shopware-cli/account-api"
)

var loginCmd = &cobra.Command{
	Use:   "account:login",
	Short: "Login into your Shopware Account",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		email := viper.GetString("account_email")
		password := viper.GetString("account_password")
		newCredentials := false

		if len(email) == 0 || len(password) == 0 {
			email, password = askUserForEmailAndPassword()
			newCredentials = true

			viper.Set("account_email", email)
			viper.Set("account_password", password)
		} else {
			termColor.Blue("Using existing credentials. Use account:logout to logout")
		}

		client, err := accountApi.NewApi(accountApi.LoginRequest{Email: email, Password: password})

		if err != nil {
			termColor.Red("Login failed with error: %s", err.Error())
			os.Exit(1)
		}

		if newCredentials {
			viper.Set("account_membership", client.GetActiveCompanyId())
			err := viper.WriteConfig()

			if err != nil {
				log.Fatalln(err)
			}
		}

		profile, err := client.GetMyProfile()

		if err != nil {
			log.Fatalln(err)
		}

		termColor.Green(
			"Hey %s %s. You are now authenticated on company %s and can use all account commands",
			profile.PersonalData.FirstName,
			profile.PersonalData.LastName,
			client.GetActiveMembership().Company.Name,
		)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func askUserForEmailAndPassword() (string, string) {
	emailPrompt := promptui.Prompt{
		Label:    "Email",
		Validate: emptyValidator,
	}

	email, err := emailPrompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	passwordPrompt := promptui.Prompt{
		Label:    "Password",
		Validate: emptyValidator,
		Mask:     '*',
	}

	password, err := passwordPrompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return email, password
}

func emptyValidator(s string) error {
	if len(s) == 0 {
		return errors.New("this cannot be empty")
	}

	return nil
}