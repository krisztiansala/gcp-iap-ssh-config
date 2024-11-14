package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	projectID    string
	instanceName string
	zone         string
	forceUpdate  bool
	dryRun       bool
	configFile   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "setup-ssh-config",
		Short: "Setup SSH config for GCP VM with IAP",
		Run: func(cmd *cobra.Command, args []string) {
			if projectID == "" || instanceName == "" || zone == "" {
				fmt.Println("Please provide all required arguments:")
				fmt.Println("-p, --project <your-project-id>")
				fmt.Println("-i, --instance <your-instance-name>")
				fmt.Println("-z, --zone <your-zone>")
				os.Exit(1)
			}

			_, sshOptions := getSSHCommand(projectID, instanceName, zone)
			updateSSHConfig(sshOptions)
		},
	}

	rootCmd.Flags().StringVarP(&projectID, "project", "p", "", "GCP project ID")
	rootCmd.Flags().StringVarP(&instanceName, "instance", "i", "", "GCP instance name")
	rootCmd.Flags().StringVarP(&zone, "zone", "z", "", "GCP zone")
	rootCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Force update existing entry")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the config without modifying the SSH config file")
	rootCmd.Flags().StringVar(&configFile, "config", getUserHomeDir()+"/.ssh/config", "Path to SSH config file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func getSSHCommand(projectID, instanceName, zone string) (string, map[string]string) {
	cmd := exec.Command("gcloud", "compute", "ssh", instanceName,
		"--tunnel-through-iap",
		"--dry-run",
		"--zone", zone,
		"--project", projectID)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting SSH command: %v\n Make sure the input arguments are correct and you have access to the instance!", err)
		return "", nil
	}

	// Parse SSH command output to extract options
	sshCmd := strings.TrimSpace(string(output))
	options := make(map[string]string)

	// Extract IdentityFile from -i option
	iPattern := regexp.MustCompile(`-i\s+([^\s]+)`)
	if matches := iPattern.FindStringSubmatch(sshCmd); len(matches) > 1 {
		options["IdentityFile"] = strings.Trim(matches[1], "\"'")
	}

	// Parse -o options
	parts := strings.Split(sshCmd, " -o ")
	for i, part := range parts {
		if i == 0 {
			continue // Skip the initial ssh command part
		}
		// Handle the last part which might contain the username@host
		if i == len(parts)-1 {
			if idx := strings.Index(part, " "); idx != -1 {
				part = part[:idx]
			}
		}
		if keyVal := strings.SplitN(part, "=", 2); len(keyVal) == 2 {
			key := strings.TrimSpace(keyVal[0])
			value := strings.Trim(keyVal[1], "\"'")
			options[key] = value
		}
	}

	return sshCmd, options
}

func updateSSHConfig(sshOptions map[string]string) {
	if sshOptions == nil {
		fmt.Println("Failed to parse SSH command options")
		return
	}

	configPath := configFile
	hostAlias := fmt.Sprintf("compute.%s", instanceName)

	// Start with the Host line
	configLines := []string{fmt.Sprintf("Host %s", hostAlias)}
	configLines = append(configLines, fmt.Sprintf("  HostName %s", hostAlias))

	// Add all options from the SSH command
	for key, value := range sshOptions {
		key = strings.Trim(key, "\"")
		configLines = append(configLines, fmt.Sprintf("  %s %s", key, value))
	}

	// Add user if not present in options
	if _, hasUser := sshOptions["User"]; !hasUser {
		configLines = append(configLines, fmt.Sprintf("  User %s", os.Getenv("USER")))
	}

	configContent := strings.Join(configLines, "\n")

	if dryRun {
		fmt.Printf("The following configuration would be added to %s:\n\n", configPath)
		fmt.Println(configContent)
		fmt.Printf("\nTo add this configuration manually, append the above content to %s\n", configPath)
		return
	}

	// Read existing config
	existingConfig, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error reading SSH config file: %v\n", err)
		return
	}

	// Check if entry exists
	sections := strings.Split(string(existingConfig), "\n\n")
	entryExists := false

	for _, section := range sections {
		if strings.HasPrefix(strings.TrimSpace(section), "Host "+hostAlias) {
			entryExists = true
			break
		}
	}

	if entryExists && !forceUpdate {
		fmt.Printf("SSH config entry already exists for %s. Use --force to update.\n", hostAlias)
		return
	}

	var newSections []string
	if forceUpdate {
		// Remove existing config for this instance if found
		for _, section := range sections {
			if !strings.HasPrefix(strings.TrimSpace(section), "Host "+hostAlias) {
				if strings.TrimSpace(section) != "" {
					newSections = append(newSections, section)
				}
			}
		}
	} else {
		// Keep all existing sections
		newSections = sections
	}

	// Add the new config
	newSections = append(newSections, configContent)

	// Write the updated config back to file
	finalConfig := strings.Join(newSections, "\n\n") + "\n"
	err = os.WriteFile(configPath, []byte(finalConfig), 0644)
	if err != nil {
		fmt.Printf("Error writing to SSH config file: %v\n", err)
		return
	}

	if entryExists {
		fmt.Printf("SSH config updated successfully for instance: %s\n", hostAlias)
	} else {
		fmt.Printf("SSH config added successfully for instance: %s\n", hostAlias)
	}
}

func getUserHomeDir() string {
	homedir, found := syscall.Getenv("HOME")
	if found {
		return homedir
	}
	homedir, err := os.UserHomeDir()
	if err == nil {
		return homedir
	} else {
		fmt.Printf("Error getting user home directory: %v\n", err)
	}
	return ""
}
