package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/spf13/cobra"
)

const (
	MIN_PASSPHRASE_LEN = 6
)

func addRequiredKeyDirFlag(cmd *cobra.Command) {
	cmd.Flags().String("keydir", "./keydir", "--keydir=/.bl_keystore")
	cmd.MarkFlagRequired("keydir")
}

func getPassPhrase() (string, error) {
	_, err := os.Stdout.Write([]byte("Enter a pass phrase ( min len 6 ):\n"))
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(os.Stdout)
	var passphrase string
	for scanner.Scan() {
		if scanner.Err() != nil {
			return "", scanner.Err()
		}
		passphrase = scanner.Text()
		if len(strings.TrimSpace(passphrase)) < MIN_PASSPHRASE_LEN {
			os.Stdout.Write([]byte("The minimum len should be not less then 6\n"))
			continue
		}
		break
	}
	return passphrase, nil
}

// Create account command
func addAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new-account",
		Short: "Create account",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				keydir, _ = cmd.Flags().GetString("keydir")
			)

			passphrase, err := getPassPhrase()
			if err != nil {
				log.Fatal(err)
			}

			ks := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)
			acc, err := ks.NewAccount(passphrase)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("The account successfully created: %s\n", acc.Address.Hex())
		},
	}
	addRequiredKeyDirFlag(cmd)
	return cmd
}

// Get PK key
func addGetPKCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get-pk",
		Short: "Get pk key",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				keyfile, _ = cmd.Flags().GetString("keyfile")
			)
			passphrase, err := getPassPhrase()
			if err != nil {
				log.Fatal(err)
			}
			json, err := os.ReadFile(keyfile)
			if err != nil {
				log.Fatal(err)
			}
			key, err := keystore.DecryptKey(json, passphrase)
			if err != nil {
				log.Fatal(err)
			}
			spew.Dump(key)
		},
	}

	cmd.Flags().String("keyfile", "the account's key file from the key store directory", "--keyfile=./UTC-02...json")
	cmd.MarkFlagRequired("keyfile")
	return cmd
}

func addWalletCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "wallet",
		Short: "Wallet command. Create account: 'new-account' ",
	}
	cmd.AddCommand(addAccountCmd())
	cmd.AddCommand(addGetPKCmd())
	return cmd
}
