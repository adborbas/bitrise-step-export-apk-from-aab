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

// Bundletooler ...
type Bundletooler interface {
	BuildAPKs(aabPath string, keystoreCfg *bundletool.KeystoreConfig) (string, error)
}

// Exporter ...
type Exporter struct {
	bundletooler Bundletooler
}

// New ...
func New(bundletooler Bundletooler) Exporter {
	return Exporter{bundletooler: bundletooler}
}

// unzipUniversalAPKsArchive unzips an universal apks archive.
func unzipUniversalAPKsArchive(archive, destDir string) (string, error) {
	unzipCommand := command.New("unzip", archive, "-d", destDir)
	return filepath.Join(destDir, "universal.apk"), run(unzipCommand)
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

// ExportUniversalAPK generates universal apks from an aab file.
func (exporter Exporter) ExportUniversalAPK(aabPath string, keystoreCfg *bundletool.KeystoreConfig) (string, error) {
	apksPth, err := exporter.bundletooler.BuildAPKs(aabPath, keystoreCfg)
	if err != nil {
		return "", err
	}

	tmpPth, err := pathutil.NormalizedOSTempDirPath("universal-apk")
	if err != nil {
		return "", err
	}
	universalAPKPath, err := unzipUniversalAPKsArchive(apksPth, tmpPth)
	if err != nil {
		return "", err
	}

	renamedUniversalAPKPath := filepath.Join(tmpPth, androidartifact.UniversalAPKBase(aabPath))
	return renamedUniversalAPKPath, os.Rename(universalAPKPath, renamedUniversalAPKPath)
}
