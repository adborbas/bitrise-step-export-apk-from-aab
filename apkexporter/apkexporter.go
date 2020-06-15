package apkexporter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adborbas/bitrise-step-export-apk-from-aab/bundletool"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/steps-deploy-to-bitrise-io/androidartifact"
)

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

// buildApksArchive generates universal apks from an aab file.
func buildApksArchive(bundleTool bundletool.Path, tmpPth, aabPth string, keystoreCfg *bundletool.KeystoreConfig) (string, error) {
	pth := filepath.Join(tmpPth, "universal.apks")
	args := []string{}
	args = append(args, "--mode=universal")
	args = append(args, "--bundle", aabPth)
	args = append(args, "--output", pth)

	if keystoreCfg != nil {
		args = append(args, "--ks", keystoreCfg.Path)
		args = append(args, "--ks-pass", keystoreCfg.KeystorePassword)
		args = append(args, "--ks-key-alias", keystoreCfg.SigningKeyAlias)
		args = append(args, "--key-pass", keystoreCfg.SigningKeyPassword)
	}

	buildAPKsCommand := bundleTool.Command("build-apks", args...)
	return pth, run(buildAPKsCommand)
}

// unzipUniversalAPKsArchive unzips an universal apks archive.
func unzipUniversalAPKsArchive(archive, destDir string) (string, error) {
	unzipCommand := command.New("unzip", archive, "-d", destDir)
	return filepath.Join(destDir, "universal.apk"), run(unzipCommand)
}

// GenerateUniversalAPK generates universal apks from an aab file.
func GenerateUniversalAPK(aabPth, bundletoolVersion string, keystoreCfg *bundletool.KeystoreConfig) (string, error) {
	bundletoolPath, err := bundletool.New(bundletoolVersion)
	if err != nil {
		return "", err
	}

	tmpPth, err := pathutil.NormalizedOSTempDirPath("aab-bundle")
	if err != nil {
		return "", err
	}

	apksPth, err := buildApksArchive(bundletoolPath, tmpPth, aabPth, keystoreCfg)
	if err != nil {
		return "", err
	}

	universalAPKPath, err := unzipUniversalAPKsArchive(apksPth, tmpPth)
	if err != nil {
		return "", err
	}

	renamedUniversalAPKPath := filepath.Join(tmpPth, androidartifact.UniversalAPKBase(aabPth))
	return renamedUniversalAPKPath, os.Rename(universalAPKPath, renamedUniversalAPKPath)
}
