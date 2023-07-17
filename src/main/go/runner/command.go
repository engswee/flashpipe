package runner

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

// TODO - to be removed when fully ported to Go

func constructClassPath(prefix string, flashpipeLocation string) (string, error) {
	paths := []string{
		"/org/codehaus/groovy/groovy-all/2.4.21/groovy-all-2.4.21.jar",
		"/org/apache/httpcomponents/core5/httpcore5/5.0.4/httpcore5-5.0.4.jar",
		"/org/apache/httpcomponents/client5/httpclient5/5.0.4/httpclient5-5.0.4.jar",
		"/commons-codec/commons-codec/1.15/commons-codec-1.15.jar",
		"/org/slf4j/slf4j-api/1.7.25/slf4j-api-1.7.25.jar",
		"/org/apache/logging/log4j/log4j-slf4j-impl/2.17.1/log4j-slf4j-impl-2.17.1.jar",
		"/org/apache/logging/log4j/log4j-api/2.17.1/log4j-api-2.17.1.jar",
		"/org/apache/logging/log4j/log4j-core/2.17.1/log4j-core-2.17.1.jar",
		"/org/zeroturnaround/zt-zip/1.14/zt-zip-1.14.jar",
	}
	var builder strings.Builder
	for _, path := range paths {
		_, err := builder.WriteString(prefix + path + ":")
		if err != nil {
			return "", err
		}
	}
	_, err := builder.WriteString(flashpipeLocation)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func JavaCmd(className string, mavenRepoPrefix string, flashpipeLocation string, log4jFile string) (string, error) {
	classPath, err := constructClassPath(mavenRepoPrefix, flashpipeLocation)
	if err != nil {
		log.Error().Msg(err.Error())
		return "", err
	}
	var cmd *exec.Cmd
	if log4jFile == "" {
		log.Info().Msgf("Executing command: java -classpath %v %v", classPath, className)
		cmd = exec.Command("java", "-classpath", classPath, className)
	} else {
		logConfig := "-Dlog4j.configurationFile=" + log4jFile
		log.Info().Msgf("Executing command: java %v -classpath %v %v", logConfig, classPath, className)
		cmd = exec.Command("java", logConfig, "-classpath", classPath, className)
	}

	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	fmt.Println(output)

	return output, err
}

func JavaCmdWithArgs(mavenRepoPrefix string, flashpipeLocation string, log4jFile string, args ...string) (string, error) {
	classPath, err := constructClassPath(mavenRepoPrefix, flashpipeLocation)
	if err != nil {
		log.Error().Msg(err.Error())
		return "", err
	}
	var cmd *exec.Cmd
	if log4jFile == "" {
		fullArgs := []string{"-classpath", classPath}
		fullArgs = append(fullArgs, args...)
		argsAny := []any{"Executing command: java", "-classpath", classPath}
		for _, arg := range args {
			argsAny = append(argsAny, arg)
		}
		//log.Info().Msgf(argsAny...)
		cmd = exec.Command("java", fullArgs...)
	} else {
		logConfig := "-Dlog4j.configurationFile=" + log4jFile
		fullArgs := []string{logConfig, "-classpath", classPath}
		fullArgs = append(fullArgs, args...)
		argsAny := []any{"Executing command: java", logConfig, "-classpath", classPath}
		for _, arg := range args {
			argsAny = append(argsAny, arg)
		}
		//log.Info().Msg(argsAny...)
		cmd = exec.Command("java", fullArgs...)
	}

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Println(string(stdoutStderr))

	return string(stdoutStderr), err
}
