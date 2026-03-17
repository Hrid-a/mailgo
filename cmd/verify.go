/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Hrid-a/mailgo/internal/verifier"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify [email]",
	Short: "Verify if an email address is valid and deliverable",
	Long: `verify runs a full check on the given email address:              
                                                                               
    1. Syntax validation      — is the format valid?                           
    2. DNS / MX lookup        — does the domain have mail servers?             
    3. SMTP handshake         — does this specific mailbox exist?              
    4. Catch-all detection    — does the server accept everything?             
    5. Disposable check       — is it a throwaway address?                     
    6. Role-based check       — is it admin@, info@, noreply@?                 
                                                                               
  Returns a verdict: deliverable | undeliverable | risky | unknown             
                                                                               
  Examples:                                                                    
    mailgo verify user@example.com
    mailgo verify user@example.com --json                              
    mailgo verify --file emails.txt --concurrency 10
    mailgo verify --file emails.txt --output results.json`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")

		output, _ := cmd.Flags().GetString("output")
		asJSON, _ := cmd.Flags().GetBool("json")

		var opts []verifier.Option

		if asJSON {
			opts = append(opts, verifier.WithJSONOutput())
		}

		if output != "" {
			opts = append(opts, verifier.WithOutputFile(output))
		}

		if file != "" {
			opts = append(opts, verifier.WithEmailsFromFile(file))
		} else if len(args) > 0 {
			opts = append(opts, verifier.WithEmailArg(args[0]))
		} else {
			fmt.Fprintln(os.Stderr, "Error: provide an email argument or --file")
			os.Exit(1)
		}

		v, err := verifier.NewVerifier(opts...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		err = v.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	},
}

// verifyCmd.Flags().IntP("concurrency", "c", 5, "Number of concurrent
//  workers for bulk verification")

func init() {
	rootCmd.AddCommand(verifyCmd)

	// verifyCmd.Flags().StringP("file", "f", "", "Path to a .txt or .csv file containing emails to verify")
	verifyCmd.Flags().StringP("output", "o", "", "Path to write results to ex: results.json")
	verifyCmd.Flags().Bool("json", false, "Output results in JSON format")
}
