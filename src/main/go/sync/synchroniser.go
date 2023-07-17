package sync

import (
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/str"
	"github.com/rs/zerolog/log"
	"os"
)

type Synchroniser struct {
	exe *httpclnt.HTTPExecuter
	ip  *odata.IntegrationPackage
}

func New(exe *httpclnt.HTTPExecuter) *Synchroniser {
	s := new(Synchroniser)
	s.exe = exe
	s.ip = odata.NewIntegrationPackage(exe)
	return s
}

func (s *Synchroniser) SyncPackageDetails(packageId string) {
	log.Info().Msgf("Processing details of integration package %v", packageId)
	readOnly, err := s.ip.IsReadOnly(packageId)
	logger.ExitIfError(err)
	if readOnly {
		log.Warn().Msgf("Skipping package %v as it is Configure-only", packageId)
		return
	}
	// TODO - complete sync package details
}

func (s *Synchroniser) SyncArtifacts(packageId string, workDir string, gitSrcDir string, includedIds []string, excludedIds []string, draftHandling string, dirNamingType string, normaliseManifestAction string, normaliseManifestPrefixOrSuffix string) {

	// Verify the package is downloadable
	readOnly, err := s.ip.IsReadOnly(packageId)
	logger.ExitIfError(err)
	if readOnly {
		log.Warn().Msgf("Skipping package %v as it is Configure-only and cannot be downloaded", packageId)
		return
	}

	// Get all design time artifacts of package
	log.Info().Msgf("Getting artifacts in integration package %v", packageId)
	artifacts, err := s.ip.GetAllArtifacts(packageId)
	logger.ExitIfError(err)

	// Create temp directories in working dir
	err = os.MkdirAll(workDir+"/download", os.ModePerm)
	logger.ExitIfError(err)
	// TODO - collect error for handling
	//err = os.MkdirAll(workDir+"/from_git", os.ModePerm)
	//logger.ExitIfError(err)
	//err = os.MkdirAll(workDir+"/from_tenant", os.ModePerm)
	//logger.ExitIfError(err)

	filtered, err := filterArtifacts(artifacts, includedIds, excludedIds)
	logger.ExitIfError(err)

	// Process through the artifacts
	for _, artifact := range filtered {
		log.Info().Msg("---------------------------------------------------------------------------------")
		log.Info().Msgf("📢 Begin processing for artifact %v", artifact.Id)
		// Check if artifact is in draft version
		if artifact.IsDraft {
			switch draftHandling {
			case "SKIP":
				log.Warn().Msgf("Artifact %v is in draft version, and will be skipped", artifact.Id)
				continue
			case "ADD":
				log.Info().Msgf("Artifact %v is in draft version, and will be added", artifact.Id)
			case "ERROR":
				logger.ExitIfError(fmt.Errorf("Artifact %v is in draft version. Save Version in Web UI first!", artifact.Id))
			}
		}
		// Download artifact content
		dt := odata.NewDesigntimeArtifact(artifact.ArtifactType, s.exe)
		targetDownloadFile := fmt.Sprintf("%v/download/%v.zip", workDir, artifact.Id)
		err = odata.Download(targetDownloadFile, artifact.Id, dt)
		logger.ExitIfError(err)

		// Normalise ID and Name
		normalisedId := str.Normalise(artifact.Id, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		normalisedName := str.Normalise(artifact.Name, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		log.Debug().Msgf("Normalised artifact ID - %v", normalisedId)
		log.Debug().Msgf("Normalised artifact name - %v", normalisedName)

		var directoryName string
		if dirNamingType == "NAME" {
			directoryName = normalisedName
		} else {
			directoryName = normalisedId
		}
		// Unzip artifact contents
		log.Debug().Msgf("Target artifact directory name - %v", directoryName)
		downloadedArtifactPath := fmt.Sprintf("%v/download/%v", workDir, directoryName)
		err = file.UnzipSource(targetDownloadFile, downloadedArtifactPath)
		logger.ExitIfError(err)
		log.Info().Msgf("Downloaded artifact unzipped to %v", downloadedArtifactPath)

		// Normalise MANIFEST.MF before sync to Git - TODO
		// https://github.com/gnewton/jargo/blob/master/jar.go
		//https://pkg.go.dev/github.com/quay/claircore/java/jar
		//https://github.com/quay/claircore/blob/v1.5.8/java/jar/jar.go
		//https://pkg.go.dev/net/textproto#Reader.ReadMIMEHeader

		//ScriptCollection scriptCollection = ScriptCollection.newInstance(scriptCollectionMap)
		//Map collections = scriptCollection.getCollections()
		//ManifestHandler.newInstance("${workDir}/download/${directoryName}/META-INF/MANIFEST.MF").normalizeAttributesInFile(normalizedIFlowID, normalizedIFlowName, scriptCollection.getTargetCollectionValues())

		// Normalise the script collection in IFlow BPMN2 XML before syncing to Git - TODO
		//if (collections.size()) {
		//	BPMN2Handler bpmn2Handler = new BPMN2Handler()
		//	bpmn2Handler.updateFiles(collections, "${workDir}/download/${directoryName}")
		//}

		gitArtifactPath := fmt.Sprintf("%v/%v", gitSrcDir, directoryName)
		if file.CheckFileExists(fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitArtifactPath)) {
			// (1) If IFlow already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")

			// TODO - no longer required?
			// Copy to temp directory for diff comparison
			// Remove comments from parameters.prop before comparison only if it exists

			// Diff directories excluding parameters.prop
			dirDiffer := file.DiffDirectories(downloadedArtifactPath, gitArtifactPath)
			// Diff parameters.prop ignoring commented lines
			downloadedParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", downloadedArtifactPath)
			gitParams := fmt.Sprintf("%v/src/main/resources/parameters.prop", gitArtifactPath)
			var paramDiffer bool
			if file.CheckFileExists(downloadedParams) && file.CheckFileExists(gitParams) {
				paramDiffer = file.DiffParams(downloadedParams, gitParams)
			} else if !file.CheckFileExists(downloadedParams) && !file.CheckFileExists(gitParams) {
				log.Warn().Msg("Skipping diff of parameters.prop as it does not exist in both source and target")
			} else {
				paramDiffer = true
				log.Info().Msg("Update required since parameters.prop does not exist in either source or target")
			}

			if dirDiffer || paramDiffer {
				log.Info().Msg("🏆 Changes detected and will be updated to Git")
				// Update the changes into the Git directory
				err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
				logger.ExitIfError(err)
			} else {
				log.Info().Msg("🏆 No changes detected. Update to Git not required")
			}

		} else { // (2) If IFlow does not exist in Git, then add it
			log.Info().Msgf("🏆 Artifact %v does not exist, and will be added to Git", artifact.Id)

			err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
			logger.ExitIfError(err)
		}
	}

	// TODO - write error wrapper - https://go.dev/blog/errors-are-values
	// Clean up working directory
	err = os.RemoveAll(workDir + "/download")
	logger.ExitIfError(err)
	//err = os.RemoveAll(workDir + "/from_git")
	//logger.ExitIfError(err)
	//err = os.RemoveAll(workDir + "/from_tenant")
	//logger.ExitIfError(err)

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of artifacts in integration package %v", packageId)
}

func filterArtifacts(artifacts []*odata.ArtifactDetails, includedIds []string, excludedIds []string) ([]*odata.ArtifactDetails, error) {
	var output []*odata.ArtifactDetails
	if len(includedIds) > 0 {
		for _, id := range includedIds {
			artifact := odata.FindArtifactById(id, artifacts)
			if artifact != nil {
				output = append(output, artifact)
			} else {
				return nil, fmt.Errorf("Artifact %v in INCLUDE_IDS does not exist", id)
			}
		}
		return output, nil
	} else if len(excludedIds) > 0 {
		for _, id := range excludedIds {
			artifact := odata.FindArtifactById(id, artifacts)
			if artifact == nil {
				return nil, fmt.Errorf("Artifact %v in EXCLUDE_IDS does not exist", id)
			}
		}
		for _, artifact := range artifacts {
			if !str.Contains(artifact.Id, excludedIds) {
				output = append(output, artifact)
			}
		}
		return output, nil
	}
	return artifacts, nil
}
