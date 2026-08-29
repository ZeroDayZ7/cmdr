// cmd/crypto/random.go
package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	randomLength int
	randomType   string
)

// newRandomCmd returns the random subcommand
func newRandomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "random",
		Aliases: []string{"rand", "r"},
		Short:   "Generate random data: strings, passwords, AES keys, numbers",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch randomType {
			case "string", "s":
				result, err := crypto.GenerateRandomString(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random string:", result)

			case "password", "pass", "p":
				result, err := crypto.GenerateRandomPassword(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random password:", result)

			case "aes", "a":
				key, err := crypto.GenerateRandomAESKey(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random AES key (HEX):   ", hex.EncodeToString(key))
				fmt.Println("Random AES key (Base64):", base64.StdEncoding.EncodeToString(key))

			case "number", "num", "n":
				num, err := crypto.GenerateRandomNumber(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random number:", num)

			default:
				return fmt.Errorf("unsupported random type: %s (supported: string/s, password/p, aes/a, number/n)", randomType)
			}
			return nil
		},
	}

	// Flagi z obsługą krótkich liter: -l / -L oraz -t
	cmd.Flags().IntVarP(&randomLength, "length", "l", 16, "Length of the random value")
	cmd.Flags().StringVarP(&randomType, "type", "t", "string", "Type of random value (string, password, aes, number)")

	return cmd
}
