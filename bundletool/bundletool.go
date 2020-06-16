package bundletool

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

// KeystoreConfig ...
type KeystoreConfig struct {
	Path               string
	KeystorePassword   string
	SigningKeyAlias    string
	SigningKeyPassword string
}

// Tool ...
type Tool struct {
	path string
}

// New ...
func New(version string) (*Tool, error) {
	tmpPth, err := pathutil.NormalizedOSTempDirPath("tool")
	if err != nil {
		return nil, err
	}

	resp, err := getFromMultipleSources([]string{
		"https://github.com/google/bundletool/releases/download/" + version + "/bundletool-all-" + version + ".jar",
		"https://github.com/google/bundletool/releases/download/" + version + "/bundletool-all.jar",
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close body, error:", err)
		}
	}()

	toolPath := filepath.Join(tmpPth, "bundletool-all.jar")

	f, err := os.Create(toolPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Failed to close file, error:", err)
		}
	}()

	_, err = io.Copy(f, resp.Body)
	log.Infof("bundletool path created at: %s", toolPath)
	return &Tool{path: toolPath}, err
}

// BuildCommand ...
func (tool Tool) BuildCommand(cmd string, args ...string) *command.Model {
	return command.New("java", append([]string{"-jar", string(tool.path), cmd}, args...)...)
}

// BuildAPKs generates universal apks from an aab file.
func (tool Tool) BuildAPKs(aabPath string, keystoreCfg *KeystoreConfig) (string, error) {
	tmpPth, err := pathutil.NormalizedOSTempDirPath("aab-bundle")
	if err != nil {
		return "", err
	}

	pth := filepath.Join(tmpPth, "universal.apks")
	args := []string{}
	args = append(args, "--mode=universal")
	args = append(args, "--bundle", aabPath)
	args = append(args, "--output", pth)

	if keystoreCfg != nil {
		args = append(args, "--ks", keystoreCfg.Path)
		args = append(args, "--ks-pass", keystoreCfg.KeystorePassword)
		args = append(args, "--ks-key-alias", keystoreCfg.SigningKeyAlias)
		args = append(args, "--key-pass", keystoreCfg.SigningKeyPassword)
	}

	buildAPKsCommand := tool.BuildCommand("build-apks", args...)
	return pth, run(buildAPKsCommand)
}

func getFromMultipleSources(sources []string) (*http.Response, error) {
	for _, source := range sources {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			log.Infof("URL used to download bundletool: %s", source)
			return resp, nil
		}
	}
	return nil, fmt.Errorf("none of the sources returned 200 OK status")
}

// handleError creates error with layout: `<cmd> failed (status: <status_code>): <cmd output>`.
func handleError(cmd, out string, err error) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("%s failed", cmd)
	if status, exitCodeErr := errorutil.CmdExitCodeFromError(err); exitCodeErr == nil {
		msg += fmt.Sprintf(" (status: %d)", status)
	}
	if len(out) > 0 {
		msg += fmt.Sprintf(": %s", out)
	}
	return errors.New(msg)
}

// run executes a given command.
func run(cmd *command.Model) error {
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	return handleError(cmd.PrintableCommandArgs(), out, err)
}
