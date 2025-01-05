package installation

import (
	"fmt"
	"os"
	"os/exec"

	"alphalab-cli/pkg/setting"
)

type UpdateCmd struct {
}

func (g *UpdateCmd) Run() error {
	const folderName = setting.InstallFolder
	const repoUrl = setting.RepoUrl

	// Check if the folder exists
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		// Clone the repo if the folder doesn't exist
		fmt.Println("Cloning the repository...")
		cmd := exec.Command("git", "clone", repoUrl)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
	} else {
		// Update the repo if the folder exists
		fmt.Println("Updating the repository...")
		cmd := exec.Command("git", "-C", folderName, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update repository: %v", err)
		}
	}

	fmt.Println("Repository is up to date.")
	return nil
}
