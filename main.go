package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adborbas/bitrise-step-export-apk-from-aab/apkexporter"
	"github.com/adborbas/bitrise-step-export-apk-from-aab/bundletool"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(config)
	fmt.Println()

	bundletoolTool, err := bundletool.New("0.15.0")
	if err != nil {
		log.Errorf("Failed to initialize bundletool: %s \n", err)
		os.Exit(1)
	}

	exporter := apkexporter.New(bundletoolTool)
	apkPath, err := exporter.ExportUniversalAPK(config.AABPath, nil)
	if err != nil {
		log.Errorf("Failed to export apk, error: %s \n", err)
		os.Exit(1)
	}

	exportEnvironmentWithEnvman("APKS_PATH", apkPath)
	log.Infof("Success apk exported to: %v", apkPath)
	os.Exit(0)
}

func exportEnvironmentWithEnvman(keyStr, valueStr string) error {
	cmd := command.New("envman", "add", "--key", keyStr)
	cmd.SetStdin(strings.NewReader(valueStr))
	return cmd.Run()
}

func apkFileName(aabPath string) string {
	fileNameWithoutExtension := strings.TrimSuffix(aabPath, filepath.Ext(aabPath))
	return fileNameWithoutExtension + ".apks"
}
