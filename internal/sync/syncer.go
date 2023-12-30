package sync

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"os"
)

//type APIMSynchroniser struct {
//	exe    *httpclnt.HTTPExecuter
//	target string
//	//ip  *api.IntegrationPackage
//}

type Syncer interface {
	Exec(workDir string, artifactsDir string, includedIds []string, excludedIds []string) error
}

func NewSyncer(target string, functionType string, exe *httpclnt.HTTPExecuter) Syncer {
	switch functionType {
	case "APIM":
		switch target {
		case "local":
			return NewAPIMLocalSynchroniser(exe)
		//case "remote":
		//	return NewScriptCollection(exe)
		default:
			return nil
		}
	//case "CPI":
	//	return NewScriptCollection(exe)
	default:
		return nil
	}
}

type APIMLocalSynchroniser struct {
	exe *httpclnt.HTTPExecuter
	//typ string
}

// NewAPIMLocalSynchroniser returns an initialised APIMLocalSynchroniser instance.
func NewAPIMLocalSynchroniser(exe *httpclnt.HTTPExecuter) Syncer {
	s := new(APIMLocalSynchroniser)
	s.exe = exe
	//mm.typ = "MessageMapping"
	return s
}

func (s *APIMLocalSynchroniser) Exec(workDir string, artifactsDir string, includedIds []string, excludedIds []string) error {
	log.Info().Msg("Sync APIM content to local")

	proxy := api.NewAPIProxy(s.exe)
	// Get all APIProxies
	artifacts, err := proxy.List()
	if err != nil {
		return err
	}

	// Create temp directories in working dir
	targetRootDir := fmt.Sprintf("%v/download", workDir)
	err = os.MkdirAll(targetRootDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	// TODO
	//filtered, err := filterArtifacts(artifacts, includedIds, excludedIds)
	//if err != nil {
	//	return err
	//}

	// Process through the artifacts
	for i, artifact := range artifacts {
		log.Info().Msgf("Syncing APIProxy %v (%v/%v)", artifact.Name, i+1, len(artifacts))
		// Download artifact content
		err = proxy.Download(artifact.Name, targetRootDir)
		if err != nil {
			return err
		}

		// Unzip artifact contents - not required

		gitArtifactPath := fmt.Sprintf("%v/%v", artifactsDir, artifact.Name)
		if file.Exists(fmt.Sprintf("%v/manifest.json", gitArtifactPath)) {
			// (1) If artifact already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")

		} else { // (2) If artifact does not exist in Git, then add it
			log.Info().Msgf("🏆 Artifact %v does not exist, and will be added to Git", artifact.Name)

		}
	}

	// Clean up working directory
	err = os.RemoveAll(workDir + "/download")
	if err != nil {
		return errors.Wrap(err, 0)
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of APIProxies")

	return nil
}