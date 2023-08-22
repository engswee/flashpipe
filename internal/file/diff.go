package file

import (
	"github.com/rs/zerolog/log"
	"os/exec"
)

func DiffDirectories(firstDir string, secondDir string) bool {
	log.Info().Msgf("Executing command: diff --ignore-matching-lines=^Origin.* --strip-trailing-cr --recursive --ignore-all-space --ignore-blank-lines --exclude=parameters.prop --exclude=.DS_Store %v %v", firstDir, secondDir)
	cmd := exec.Command("diff", "--ignore-matching-lines=^Origin.*", "--strip-trailing-cr", "--recursive", "--ignore-all-space", "--ignore-blank-lines", "--exclude=parameters.prop", "--exclude=.DS_Store", firstDir, secondDir)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Info().Msgf("Diff results:\n%v", string(stdoutStderr))
	}

	return err != nil
}

func DiffFile(firstFile string, secondFile string) bool {
	// - ignoring commented lines (beginning with #)
	// - ignoring blank lines and extra white space
	log.Info().Msgf("Executing command: diff --ignore-matching-lines=^#.* --strip-trailing-cr --ignore-all-space --ignore-blank-lines %v %v", firstFile, secondFile)
	cmd := exec.Command("diff", "--ignore-matching-lines=^#.*", "--strip-trailing-cr", "--ignore-all-space", "--ignore-blank-lines", firstFile, secondFile)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Info().Msgf("Diff results:\n%v", string(stdoutStderr))
	}

	return err != nil
}
