package main

import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const name = "aws-keymaster"
const version = "0.1"

func main() {
	var dryRun bool
	importCmd := &cobra.Command{
		Use:   "import [name] [public key file]",
		Short: "Imports a public key into all AWS regions",
		Long:  "A simple utility that makes it easy to import a public key into all AWS regions with a single command",
		Run: func(cmd *cobra.Command, args []string) {
			var name string
			var pubKey string
			pubKeyDefault := guessPublicKey()

			switch len(args) {
			case 0:
				name = prompt("Key Name", "")
				pubKey = prompt("Public key", pubKeyDefault)
				fmt.Println("")
			case 1:
				name = args[0]
				pubKey = prompt("Public key", pubKeyDefault)
				fmt.Println("")
			default:
				name = strings.TrimSpace(args[0])
				pubKey = strings.TrimSpace(args[1])
			}

			if name == "" {
				fmt.Print("Key name is required.\n\n")
				cmd.Usage()
				return
			}

			if pubKey == "nil" {
				fmt.Print("Public key file is required.\n\n")
				cmd.Usage()
				return
			}

			err := importKeyPair(name, pubKey, dryRun)
			if err != nil {
				fmt.Printf("Could not import key pair: %v\n", err)
			}
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Deletes a keypair from all AWS regions",
		Long:  "Deletes a keypair with the specified name from all AWS regions",
		Run: func(cmd *cobra.Command, args []string) {
			var name string
			if len(args) < 1 {
				name = prompt("Key Name", "")
				fmt.Println("")
			} else {
				name = strings.TrimSpace(args[0])
			}

			if name == "" {
				fmt.Print("Key name is required.\n\n")
				cmd.Usage()
				return
			}

			err := deleteKeyPair(name, dryRun)
			if err != nil {
				fmt.Printf("Could not delete key pair: %v\n", err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Display the version of %s", name),
		Long:  fmt.Sprintf("Display the version of %s", name),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}

	rootCmd := &cobra.Command{Use: name}
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Checks whether you have the required permissions, without attempting the request")
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.Execute()
}

func importKeyPair(keyName string, pubKey string, dryRun bool) error {
	data, err := ioutil.ReadFile(pubKey)
	if err != nil {
		return err
	}

	for _, region := range regions() {
		label := fmt.Sprintf("%s:", region)
		svc := ec2.New(&aws.Config{Region: aws.String(region)})

		input := &ec2.ImportKeyPairInput{
			KeyName:           aws.String(keyName),
			PublicKeyMaterial: data,
			DryRun:            aws.Bool(dryRun),
		}

		resp, err := svc.ImportKeyPair(input)

		if err != nil {
			errMsg := err.Error()
			switch {
			case dryRun && strings.HasPrefix(errMsg, "DryRunOperation"):
				fmt.Printf("[Dry Run] %-16s Imported keypair '%s'\n", label, keyName)
			case strings.HasPrefix(errMsg, "InvalidKeyPair.Duplicate"):
				fmt.Printf("%-16s Keypair '%s' already exists.\n", label, keyName)
			default:
				fmt.Printf("%-16s Could not import keypair '%s' - %v\n", label, keyName, err)
			}
			continue
		}

		fmt.Printf("%-16s Imported keypair '%s' - %v\n", label, keyName, *resp.KeyFingerprint)
	}
	return nil
}

func keyPairExists(svc *ec2.EC2, keyName string, dryRun bool) bool {
	input := &ec2.DescribeKeyPairsInput{
		DryRun:   aws.Bool(dryRun),
		KeyNames: []*string{aws.String(keyName)},
	}

	resp, err := svc.DescribeKeyPairs(input)
	if err != nil {
		return false
	}

	return len(resp.KeyPairs) > 0
}

func deleteKeyPair(keyName string, dryRun bool) error {
	for _, region := range regions() {
		label := fmt.Sprintf("%s:", region)
		svc := ec2.New(&aws.Config{Region: aws.String(region)})

		exists := keyPairExists(svc, keyName, dryRun)
		if !exists {
			fmt.Printf("%-16s Keypair '%s' does not exist\n", label, keyName)
			continue
		}

		input := &ec2.DeleteKeyPairInput{
			KeyName: aws.String(keyName),
			DryRun:  aws.Bool(dryRun),
		}

		_, err := svc.DeleteKeyPair(input)

		if err != nil {
			errMsg := err.Error()
			switch {
			case dryRun && strings.HasPrefix(errMsg, "DryRunOperation"):
				fmt.Printf("[Dry Run] %-16s Deleted keypair '%s'\n", label, keyName)
			default:
				fmt.Printf("%-16s Could not delete keypair '%s' - %v\n", label, region, err)
			}
			continue
		}

		fmt.Printf("%-16s Deleted keypair '%s'\n", label, keyName)
	}
	return nil
}

func regions() []string {
	svc := ec2.New(&aws.Config{Region: aws.String("us-east-1")})

	resp, err := svc.DescribeRegions(nil)
	if err != nil {
		panic(err)
	}

	regions := make([]string, len(resp.Regions))
	for i, region := range resp.Regions {
		regions[i] = *region.RegionName
	}

	return regions
}

func prompt(name string, defaultVal string) string {
	var p string

	r := bufio.NewReader(os.Stdin)

	if defaultVal != "" {
		p = fmt.Sprintf("%s [%s]: ", name, defaultVal)
	} else {
		p = fmt.Sprintf("%s: ", name)
	}

	fmt.Print(p)

	val, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}

	val = strings.TrimSpace(val)

	if val == "" && defaultVal != "" {
		val = defaultVal
	}

	return val
}

func guessPublicKey() string {
	var pubKey string
	home, err := homedir.Dir()
	if err != nil {
		pubKey = ""
	} else {
		pubKey = path.Join(home, ".ssh/id_rsa.pub")
	}
	return pubKey
}
