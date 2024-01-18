package sync

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"slices"
)

type Syncer interface {
	Exec(workDir string, artifactsDir string, includedIds []string, excludedIds []string) error
}

func NewSyncer(target string, functionType string, exe *httpclnt.HTTPExecuter) Syncer {
	switch functionType {
	case "APIM":
		switch target {
		case "local":
			return NewAPIMLocalSynchroniser(exe)
		case "remote":
			return NewAPIMRemoteSynchroniser(exe)
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
}

// NewAPIMLocalSynchroniser returns an initialised APIMLocalSynchroniser instance.
func NewAPIMLocalSynchroniser(exe *httpclnt.HTTPExecuter) Syncer {
	s := new(APIMLocalSynchroniser)
	s.exe = exe
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

	// Process through the artifacts
	for _, artifact := range artifacts {
		log.Info().Msg("---------------------------------------------------------------------------------")
		log.Info().Msgf("📢 Begin processing for APIProxy %v", artifact.Name)

		// Filter in/out artifacts
		if skipArtifact(artifact.Name, includedIds, excludedIds) {
			continue
		}

		// Download artifact content
		err = proxy.Download(artifact.Name, targetRootDir)
		if err != nil {
			return err
		}

		// Compare content and update Git if required
		gitArtifactPath := fmt.Sprintf("%v/%v", artifactsDir, artifact.Name)
		downloadedArtifactPath := fmt.Sprintf("%v/%v", targetRootDir, artifact.Name)
		if file.Exists(fmt.Sprintf("%v/manifest.json", gitArtifactPath)) {
			// (1) If artifact already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")
			dirDiffer := file.DiffDirectories(downloadedArtifactPath, gitArtifactPath)

			if dirDiffer {
				log.Info().Msg("🏆 Changes detected and will be updated to Git")
				// Update the changes into the Git directory
				err := file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
				if err != nil {
					return err
				}
			} else {
				log.Info().Msg("🏆 No changes detected. Update to Git not required")
			}
		} else { // (2) If artifact does not exist in Git, then add it
			log.Info().Msgf("🏆 APIProxy %v does not exist, and will be added to Git", artifact.Name)
			err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
			if err != nil {
				return err
			}
		}
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of APIProxies")

	return nil
}

type APIMRemoteSynchroniser struct {
	exe *httpclnt.HTTPExecuter
}

// NewAPIMRemoteSynchroniser returns an initialised APIMRemoteSynchroniser instance.
func NewAPIMRemoteSynchroniser(exe *httpclnt.HTTPExecuter) Syncer {
	s := new(APIMRemoteSynchroniser)
	s.exe = exe
	return s
}

func (s *APIMRemoteSynchroniser) Exec(workDir string, artifactsDir string, includedIds []string, excludedIds []string) error {
	// Get directory list
	baseSourceDir := filepath.Clean(artifactsDir)
	entries, err := os.ReadDir(baseSourceDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	proxy := api.NewAPIProxy(s.exe)

	// Create temp directories in working dir
	uploadWorkDir := fmt.Sprintf("%v/upload", workDir)
	err = os.MkdirAll(uploadWorkDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	downloadWorkDir := fmt.Sprintf("%v/download", workDir)
	err = os.MkdirAll(downloadWorkDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	artifactDirFound := false
	for _, entry := range entries {
		artifactId := entry.Name()
		manifestPath := fmt.Sprintf("%v/%v/manifest.json", baseSourceDir, artifactId)
		if entry.IsDir() && file.Exists(manifestPath) {
			artifactDirFound = true
			gitArtifactDir := fmt.Sprintf("%v/%v", baseSourceDir, artifactId)

			log.Info().Msg("---------------------------------------------------------------------------------")
			log.Info().Msgf("Processing directory %v", gitArtifactDir)

			// Filter in/out artifacts
			if skipArtifact(artifactId, includedIds, excludedIds) {
				continue
			}

			log.Info().Msgf("📢 Begin processing for APIProxy %v", artifactId)
			proxyExists, err := proxy.Get(artifactId)
			if err != nil {
				return err
			}
			if !proxyExists {
				log.Info().Msgf("APIProxy %v will be created", artifactId)

				err = proxy.Upload(gitArtifactDir, uploadWorkDir)
				if err != nil {
					return err
				}

				log.Info().Msg("🏆 APIProxy created successfully")
			} else {
				log.Info().Msg("Checking if APIProxy needs to be updated")

				err = proxy.Download(artifactId, downloadWorkDir)
				if err != nil {
					return err
				}

				log.Info().Msg("Comparing content from tenant against Git")
				downloadArtifactDir := fmt.Sprintf("%v/%v", downloadWorkDir, artifactId)
				dirDiffer := file.DiffDirectories(downloadArtifactDir, gitArtifactDir)
				if dirDiffer == true {
					log.Info().Msg("Changes found in APIProxy. APIProxy will be updated in tenant")

					err = proxy.Upload(gitArtifactDir, uploadWorkDir)
					if err != nil {
						return err
					}
					log.Info().Msg("🏆 APIProxy updated successfully")
				} else {
					log.Info().Msg("🏆 No changes detected. APIProxy does not need to be updated")
				}
			}
		}
	}
	if !artifactDirFound {
		log.Warn().Msgf("No directory with APIProxy contents found in %v", baseSourceDir)
	}
	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of APIProxies")
	return nil
}

func skipArtifact(artifactId string, includedIds []string, excludedIds []string) bool {
	// Filter in/out artifacts
	if len(includedIds) > 0 {
		if !slices.Contains(includedIds, artifactId) {
			log.Warn().Msgf("Skipping %v as it is not in --ids-include", artifactId)
			return true
		}
	}
	if len(excludedIds) > 0 {
		if slices.Contains(excludedIds, artifactId) {
			log.Warn().Msgf("Skipping %v as it is in --ids-exclude", artifactId)
			return true
		}
	}
	return false
}
