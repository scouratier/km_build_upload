package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	versionFile   = "version"
	defaultVersion = "0.0.0"
	gitUserName   = "github-actions[bot]"
	gitUserEmail  = "github-actions[bot]@users.noreply.github.com"
)

// Version represents a semantic version with major, minor, and patch components.
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version in "major.minor.patch" format.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion parses a semantic version string (e.g., "1.2.3") into a Version struct.
// Returns an error if the format is invalid or any component is not a non-negative integer.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("version string is empty")
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format %q: expected 3 parts (major.minor.patch), got %d", s, len(parts))
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return Version{}, fmt.Errorf("invalid major version %q in %q: must be a non-negative integer", parts[0], s)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return Version{}, fmt.Errorf("invalid minor version %q in %q: must be a non-negative integer", parts[1], s)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return Version{}, fmt.Errorf("invalid patch version %q in %q: must be a non-negative integer", parts[2], s)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// BumpPatch increments the patch version component.
func (v Version) BumpPatch() Version {
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

// ReadVersionFile reads the version from the version file.
// If the file does not exist, returns the default version "0.0.0".
// Returns an error if the file exists but cannot be read or contains invalid content.
func ReadVersionFile(path string) (Version, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Version file %q not found, starting from %s\n", path, defaultVersion)
			return ParseVersion(defaultVersion)
		}
		return Version{}, fmt.Errorf("failed to read version file %q: %w", path, err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		fmt.Printf("Version file %q is empty, starting from %s\n", path, defaultVersion)
		return ParseVersion(defaultVersion)
	}

	v, err := ParseVersion(content)
	if err != nil {
		return Version{}, fmt.Errorf("version file %q contains invalid version: %w", path, err)
	}

	return v, nil
}

// WriteVersionFile writes the version string to the version file.
func WriteVersionFile(path string, v Version) error {
	err := os.WriteFile(path, []byte(v.String()+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write version file %q: %w", path, err)
	}
	return nil
}

// ExtractRepoName extracts the repository name from GITHUB_REPOSITORY (e.g., "owner/repo" → "repo").
func ExtractRepoName(githubRepo string) (string, error) {
	githubRepo = strings.TrimSpace(githubRepo)
	if githubRepo == "" {
		return "", fmt.Errorf("GITHUB_REPOSITORY environment variable is empty or not set")
	}

	parts := strings.Split(githubRepo, "/")
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("GITHUB_REPOSITORY %q has unexpected format: expected \"owner/repo\"", githubRepo)
	}

	return strings.ToLower(parts[1]), nil
}

// runGit executes a git command and returns its combined output.
// Provides descriptive error messages including the command that failed and its output.
func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nOutput: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

// writeGitHubOutput appends a key=value pair to the GITHUB_OUTPUT file.
// This is the official mechanism for composite actions to expose outputs to subsequent steps.
func writeGitHubOutput(key, value string) error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		// Not running in GitHub Actions — just print the output for local testing.
		fmt.Printf("[local] output %s=%s\n", key, value)
		return nil
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file %q: %w", outputFile, err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	if err != nil {
		return fmt.Errorf("failed to write to GITHUB_OUTPUT file: %w", err)
	}

	return nil
}

func run() error {
	// Step 1: Read the current version
	fmt.Println("=== Step 1: Reading version file ===")
	currentVersion, err := ReadVersionFile(versionFile)
	if err != nil {
		return fmt.Errorf("step 1 failed: %w", err)
	}
	fmt.Printf("Current version: %s\n", currentVersion)

	// Step 2: Bump the patch version
	fmt.Println("=== Step 2: Bumping patch version ===")
	newVersion := currentVersion.BumpPatch()
	fmt.Printf("New version: %s\n", newVersion)

	// Step 3: Write the new version to the file
	fmt.Println("=== Step 3: Writing new version to file ===")
	if err := WriteVersionFile(versionFile, newVersion); err != nil {
		return fmt.Errorf("step 3 failed: %w", err)
	}
	fmt.Printf("Wrote %s to %s\n", newVersion, versionFile)

	// Step 4: Configure git identity
	fmt.Println("=== Step 4: Configuring git identity ===")
	if _, err := runGit("config", "user.name", gitUserName); err != nil {
		return fmt.Errorf("step 4 failed (user.name): %w", err)
	}
	if _, err := runGit("config", "user.email", gitUserEmail); err != nil {
		return fmt.Errorf("step 4 failed (user.email): %w", err)
	}
	fmt.Printf("Git identity set to %s <%s>\n", gitUserName, gitUserEmail)

	// Step 5: Commit the version file
	fmt.Println("=== Step 5: Committing version file ===")
	commitMsg := fmt.Sprintf("chore: bump version to %s", newVersion)
	if _, err := runGit("add", versionFile); err != nil {
		return fmt.Errorf("step 5 failed (git add): %w", err)
	}
	if _, err := runGit("commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("step 5 failed (git commit): %w", err)
	}
	fmt.Printf("Committed: %s\n", commitMsg)

	// Step 6: Create the git tag
	fmt.Println("=== Step 6: Creating git tag ===")
	tag := newVersion.String()
	if _, err := runGit("tag", tag); err != nil {
		return fmt.Errorf("step 6 failed: %w\nHint: does the tag %q already exist? Check with 'git tag -l %s'", err, tag, tag)
	}
	fmt.Printf("Created tag: %s\n", tag)

	// Step 7: Push commit and tag
	fmt.Println("=== Step 7: Pushing to remote ===")
	if _, err := runGit("push", "origin", "HEAD"); err != nil {
		return fmt.Errorf("step 7 failed (push HEAD): %w\nHint: does your GITHUB_TOKEN have 'contents: write' permission?", err)
	}
	if _, err := runGit("push", "origin", tag); err != nil {
		return fmt.Errorf("step 7 failed (push tag): %w\nHint: does your GITHUB_TOKEN have 'contents: write' permission?", err)
	}
	fmt.Println("Pushed commit and tag to remote")

	// Step 8: Derive the Docker image name from the repo name
	fmt.Println("=== Step 8: Deriving Docker image name ===")
	githubRepo := os.Getenv("GITHUB_REPOSITORY")
	repoName, err := ExtractRepoName(githubRepo)
	if err != nil {
		return fmt.Errorf("step 8 failed: %w", err)
	}
	fmt.Printf("Derived image name: %s\n", repoName)

	// Step 9: Write outputs for subsequent workflow steps
	fmt.Println("=== Step 9: Writing action outputs ===")
	if err := writeGitHubOutput("new_version", newVersion.String()); err != nil {
		return fmt.Errorf("step 9 failed (new_version): %w", err)
	}
	if err := writeGitHubOutput("image_name", repoName); err != nil {
		return fmt.Errorf("step 9 failed (image_name): %w", err)
	}
	fmt.Println("Outputs written successfully")

	fmt.Printf("\n✅ Done! Version bumped from %s to %s, tagged and pushed.\n", currentVersion, newVersion)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %s\n", err)
		os.Exit(1)
	}
}
