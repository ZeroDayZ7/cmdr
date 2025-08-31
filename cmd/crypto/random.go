package crypto

import (
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
		Use:   "random",
		Short: "Generate random data: strings, passwords, AES keys, numbers",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch randomType {
			case "string":
				result, err := crypto.GenerateRandomString(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random string:", result)
			case "aes":
				key, err := crypto.GenerateRandomAESKey(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random AES key:", string(key))
			case "number":
				num, err := crypto.GenerateRandomNumber(randomLength)
				if err != nil {
					return err
				}
				fmt.Println("Random number:", num)
			default:
				return fmt.Errorf("unsupported random type: %s", randomType)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&randomLength, "length", 16, "Length of the random value")
	cmd.Flags().StringVar(&randomType, "type", "string", "Type of random value (string, aes, number)")

	return cmd
}
