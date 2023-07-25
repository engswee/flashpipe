package sync

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/file"
	"github.com/engswee/flashpipe/httpclnt"
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

func (s *Synchroniser) SyncPackageDetails(packageId string, workDir string, gitSrcDir string) error {
	packageFromTenant, readOnly, packageExists, err := s.ip.Get(packageId)
	if err != nil {
		return err
	}
	if !packageExists {
		return fmt.Errorf("Package %v does not exist", packageId)
	}
	if readOnly {
		log.Warn().Msgf("Skipping package %v as it is Configure-only", packageId)
		return nil
	}

	// Create temp directory in working dir
	err = os.MkdirAll(workDir+"/from_tenant", os.ModePerm)
	if err != nil {
		return err
	}

	// TODO - normalise ID and name - no longer required?
	// Normalize ID
	// Normalize Name

	log.Info().Msg("Storing package details from tenant for comparison")
	// Write package details from tenant to file
	tenantFile := fmt.Sprintf("%v/from_tenant/%v.json", workDir, packageId)
	f, err := os.Create(tenantFile)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := json.MarshalIndent(packageFromTenant, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}

	// Get existing package details file if it exists and compare values
	gitSourceFile := fmt.Sprintf("%v/%v.json", gitSrcDir, packageId)
	if file.CheckFileExists(gitSourceFile) {
		packageFromGit, err := odata.GetPackageDetails(tenantFile)
		if err != nil {
			return err
		}
		// TODO - Use Unix diff instead?
		if contentDiffer(packageFromTenant, packageFromGit) {
			log.Info().Msgf("🏆 Changes to package %v detected and will be updated to Git", packageId)
			err = file.CopyFile(tenantFile, gitSourceFile)
			if err != nil {
				return err
			}
		} else {
			log.Info().Msgf("🏆 No changes to package %v detected. Update to Git not required", packageId)
		}
	} else {
		log.Info().Msgf("🏆 Saving new file for package %v to Git", packageId)
		err = file.CopyFile(tenantFile, gitSourceFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Synchroniser) SyncArtifacts(packageId string, workDir string, gitSrcDir string, includedIds []string, excludedIds []string, draftHandling string, dirNamingType string, scriptCollectionMap string) error {
	// Verify the package is downloadable (not read only)
	_, readOnly, packageExists, err := s.ip.Get(packageId)
	if err != nil {
		return err
	}
	if !packageExists {
		return fmt.Errorf("Package %v does not exist", packageId)
	}
	if readOnly {
		log.Warn().Msgf("Skipping package %v as it is Configure-only and cannot be downloaded", packageId)
		return nil
	}

	// Get all design time artifacts of package
	log.Info().Msgf("Getting artifacts in integration package %v", packageId)
	artifacts, err := s.ip.GetAllArtifacts(packageId)
	if err != nil {
		return err
	}

	// Create temp directories in working dir
	err = os.MkdirAll(workDir+"/download", os.ModePerm)
	if err != nil {
		return err
	}
	// TODO - collect error for handling
	//err = os.MkdirAll(workDir+"/from_git", os.ModePerm)
	//logger.ExitIfError(err)
	//err = os.MkdirAll(workDir+"/from_tenant", os.ModePerm)
	//logger.ExitIfError(err)

	filtered, err := filterArtifacts(artifacts, includedIds, excludedIds)
	if err != nil {
		return err
	}

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
				return fmt.Errorf("Artifact %v is in draft version. Save Version in Web UI first!", artifact.Id)
			}
		}
		// Download artifact content
		dt := odata.NewDesigntimeArtifact(artifact.ArtifactType, s.exe)
		targetDownloadFile := fmt.Sprintf("%v/download/%v.zip", workDir, artifact.Id)
		err = odata.Download(targetDownloadFile, artifact.Id, dt)
		if err != nil {
			return err
		}

		// Normalise ID and Name
		//normalisedId := str.Normalise(artifact.Id, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		//normalisedName := str.Normalise(artifact.Name, normaliseManifestAction, normaliseManifestPrefixOrSuffix)
		//log.Debug().Msgf("Normalised artifact ID - %v", normalisedId)
		//log.Debug().Msgf("Normalised artifact name - %v", normalisedName)

		var directoryName string
		if dirNamingType == "NAME" {
			directoryName = artifact.Name
		} else {
			directoryName = artifact.Id
		}
		// Unzip artifact contents
		log.Debug().Msgf("Target artifact directory name - %v", directoryName)
		downloadedArtifactPath := fmt.Sprintf("%v/download/%v", workDir, directoryName)
		err = file.UnzipSource(targetDownloadFile, downloadedArtifactPath)
		if err != nil {
			return err
		}
		log.Info().Msgf("Downloaded artifact unzipped to %v", downloadedArtifactPath)

		// TODO - Normalise MANIFEST.MF before sync to Git

		// Normalise the script collection in IFlow BPMN2 XML before syncing to Git
		err = file.UpdateBPMN(downloadedArtifactPath, scriptCollectionMap)
		if err != nil {
			return err
		}

		gitArtifactPath := fmt.Sprintf("%v/%v", gitSrcDir, directoryName)
		if file.CheckFileExists(fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitArtifactPath)) {
			// (1) If IFlow already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")

			// TODO - no longer required?
			// Copy to temp directory for diff comparison
			// Remove comments from parameters.prop before comparison only if it exists

			// Diff directories excluding parameters.prop
			// TODO - diff meta and resources/value-mapping instead of whole directory
			dirDiffer := dt.DiffContent(downloadedArtifactPath, gitArtifactPath)

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
				if err != nil {
					return err
				}
			} else {
				log.Info().Msg("🏆 No changes detected. Update to Git not required")
			}

		} else { // (2) If IFlow does not exist in Git, then add it
			log.Info().Msgf("🏆 Artifact %v does not exist, and will be added to Git", artifact.Id)

			err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
			if err != nil {
				return err
			}
		}
	}

	// TODO - write error wrapper - https://go.dev/blog/errors-are-values
	// Clean up working directory
	err = os.RemoveAll(workDir + "/download")
	if err != nil {
		return err
	}
	//err = os.RemoveAll(workDir + "/from_git")
	//logger.ExitIfError(err)
	//err = os.RemoveAll(workDir + "/from_tenant")
	//logger.ExitIfError(err)

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of artifacts in integration package %v", packageId)
	return nil
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

func contentDiffer(source *odata.PackageSingleData, target *odata.PackageSingleData) bool {
	if source.Root.Name != target.Root.Name {
		return true
	}
	if source.Root.Description != target.Root.Description {
		return true
	}
	if source.Root.ShortText != target.Root.ShortText {
		return true
	}
	if source.Root.Version != target.Root.Version {
		return true
	}
	if source.Root.Vendor != target.Root.Vendor {
		return true
	}
	if source.Root.Mode != target.Root.Mode {
		return true
	}
	if source.Root.Products != target.Root.Products {
		return true
	}
	if source.Root.Keywords != target.Root.Keywords {
		return true
	}
	if source.Root.Countries != target.Root.Countries {
		return true
	}
	if source.Root.Industries != target.Root.Industries {
		return true
	}
	if source.Root.LineOfBusiness != target.Root.LineOfBusiness {
		return true
	}
	return false
}
