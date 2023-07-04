package file

import (
	"fmt"
	"github.com/engswee/flashpipe/logger"
	"os/exec"
)

func DiffDirectories(firstDir string, secondDir string) bool {
	// Any configured value will remain in IFlow even if the IFlow is replaced and the parameter is no longer used
	// Therefore diff of parameters.prop may come up with false differences
	logger.Info("Executing command:", "diff", "--ignore-matching-lines=^Origin.*", "--strip-trailing-cr", "--recursive", "--ignore-all-space", "--ignore-blank-lines", "--exclude=parameters.prop", "--exclude=metainfo.prop", "--exclude=.DS_Store", firstDir, secondDir)
	cmd := exec.Command("diff", "--ignore-matching-lines=^Origin.*", "--strip-trailing-cr", "--recursive", "--ignore-all-space", "--ignore-blank-lines", "--exclude=parameters.prop", "--exclude=metainfo.prop", "--exclude=.DS_Store", firstDir, secondDir)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
	}

	return err != nil
}

func DiffParams(firstFile string, secondFile string) bool {
	// Compare parameters.prop
	// - ignoring commented lines (beginning with #)
	// - ignorring blank lines and extra white space
	logger.Info("Executing command:", "diff", "--ignore-matching-lines=^#.*", "--strip-trailing-cr", "--ignore-all-space", "--ignore-blank-lines", firstFile, secondFile)
	cmd := exec.Command("diff", "--ignore-matching-lines=^#.*", "--strip-trailing-cr", "--ignore-all-space", "--ignore-blank-lines", firstFile, secondFile)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
	}

	return err != nil
}
