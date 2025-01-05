package installation

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"alphalab-cli/pkg/setting"

	"github.com/manifoldco/promptui"
)

// InstallCmd struct holds the repository URL needed for installation
type InstallCmd struct {
}

// Run executes the installation process
func (g *InstallCmd) Run() error {
	fmt.Println("Welcome to the AlphaLab installation process!")

	// Check if all necessary tools are installed
	if err := checkRequirements(); err != nil {
		return fmt.Errorf("failed to check requirements: %v", err)
	}

	// Prompt the user for admin email and password with validation
	email, err := promptEmail()
	if err != nil {
		return fmt.Errorf("email input error: %v", err)
	}
	password, err := promptPassword()
	if err != nil {
		return fmt.Errorf("password input error: %v", err)
	}
	if _, err := promptVerifyPassword(password); err != nil {
		return fmt.Errorf("password verification error: %v", err)
	}

	// Confirmation loop for settings verification and modification
	for {
		fmt.Println("\n========== Your Settings ==========")
		fmt.Printf("Admin Email: %s\n", email)
		fmt.Printf("Admin Password: %s\n", password)
		fmt.Println("===================================")

		// Confirm settings before proceeding
		confirm := promptui.Select{
			Label: "Please confirm your settings",
			Items: []string{"Confirm and continue", "Modify settings"},
		}

		idx, _, err := confirm.Run()
		if err != nil {
			return fmt.Errorf("confirmation selection error: %v", err)
		}

		if idx == 0 {
			// User confirmed, break the loop
			break
		}

		// If user chooses to modify settings
		choice := promptui.Select{
			Label: "What would you like to change",
			Items: []string{"Email", "Password", "Back to confirmation"},
		}

		idx, _, err = choice.Run()
		if err != nil {
			return fmt.Errorf("modification selection error: %v", err)
		}

		switch idx {
		case 0: // Modify Email
			email, err = promptEmail()
			if err != nil {
				return fmt.Errorf("email modification error: %v", err)
			}
		case 1: // Modify Password
			password, err = promptPassword()
			if err != nil {
				return fmt.Errorf("password modification error: %v", err)
			}
			if _, err := promptVerifyPassword(password); err != nil {
				return fmt.Errorf("password verification error: %v", err)
			}
		case 2: // Return to confirmation
			continue
		}
	}

	// Clear screen for a cleaner installation process
	Clear()

	fmt.Println("\nStarting AlphaLab installation...")
	// Perform the installation
	if err := installAlphalab(email, password); err != nil {
		return fmt.Errorf("failed to install AlphaLab: %v", err)
	}

	// Verify the installation process
	fmt.Println("Verifying AlphaLab installation...")
	if err := verifyAlphalabStatus(); err != nil {
		return fmt.Errorf("failed to verify AlphaLab installation: %v", err)
	}

	fmt.Println("AlphaLab installation completed successfully")
	return nil
}

// checkRequirements ensures all necessary tools are installed before proceeding
func checkRequirements() error {
	// Verify Docker is installed
	_, err := exec.Command("docker", "--version").Output()
	if err != nil {
		return fmt.Errorf("docker is not installed, please install docker first")
	}

	// Verify Docker Compose is installed
	_, err = exec.Command("docker", "compose", "version").Output()
	if err != nil {
		return fmt.Errorf("docker compose is not installed, please install docker compose first")
	}

	// Verify Git is installed
	_, err = exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Errorf("git is not installed, please install git first")
	}

	return nil
}

// installAlphalab performs the cloning and setup of the repository using Docker Compose
func installAlphalab(ADMIN_EMAIL string, ADMIN_PASSWORD string) error {
	const folderName = "AlphaLab"

	// Check if the AlphaLab directory already exists
	if _, err := os.Stat(folderName); !os.IsNotExist(err) {
		fmt.Println("The AlphaLab directory already exists.")
		// Ask user if they want to update the existing repo
		prompt := promptui.Select{
			Label: "Do you want to update the existing repository instead?",
			Items: []string{"Update the repository", "Cancel installation"},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("error in user selection: %v", err)
		}

		if idx == 0 {
			// Update the existing repository
			fmt.Println("Updating the existing AlphaLab repository...")
			cmd := exec.Command("git", "-C", folderName, "pull")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update repository: %v", err)
			}
			fmt.Println("Repository updated successfully.")
		} else {
			// Cancel installation if user chooses not to update
			fmt.Println("Installation cancelled.")
			return nil
		}
	} else {
		// If the folder does not exist, clone the repository
		fmt.Println("Cloning the repository...")
		cmd := exec.Command("git", "clone", setting.RepoUrl)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
		fmt.Println("Repository cloned successfully.")
	}

	// Change working directory to the cloned repository
	err := os.Chdir(folderName)
	if err != nil {
		return fmt.Errorf("failed to enter AlphaLab folder: %v", err)
	}

	// Run Docker Compose with provided admin credentials
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ADMIN_EMAIL=%s", ADMIN_EMAIL),
		fmt.Sprintf("ADMIN_PASSWORD=%s", ADMIN_PASSWORD),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run docker compose: %v", err)
	}

	fmt.Println("AlphaLab installation completed successfully!")
	return nil
}

// verifyAlphalabStatus is a placeholder for verification logic
func verifyAlphalabStatus() error {
	// Placeholder for actual service verification logic
	return nil
}

// Clear function clears the console screen based on the OS
func Clear() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		// Generic ANSI escape sequence for clearing screen
		fmt.Print("\033[H\033[2J")
	}
}

// promptEmail prompts the user for an admin email and validates it
func promptEmail() (string, error) {
	validate := func(input string) error {
		if !isValidEmail(input) {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Admin Email",
		Validate: validate,
	}

	return prompt.Run()
}

// promptPassword prompts the user for an admin password with validation
func promptPassword() (string, error) {
	validate := func(input string) error {
		if !isValidPassword(input) {
			return fmt.Errorf("password must be at least 8 characters long and include uppercase, lowercase, and a number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Admin Password",
		Mask:     '*',
		Validate: validate,
	}

	return prompt.Run()
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidPassword(password string) bool {
	length := len(password) >= 8
	upper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	lower := regexp.MustCompile(`[a-z]`).MatchString(password)
	number := regexp.MustCompile(`[0-9]`).MatchString(password)
	return length && upper && lower && number
}

func promptVerifyPassword(originalPassword string) (string, error) {
	validate := func(input string) error {
		if input != originalPassword {
			return fmt.Errorf("passwords do not match")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Verify Password",
		Mask:     '*',
		Validate: validate,
	}

	return prompt.Run()
}
