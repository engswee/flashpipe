package sync

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/engswee/flashpipe/internal/str"
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
	if file.Exists(gitSourceFile) {
		packageFromGit, err := odata.GetPackageDetails(tenantFile)
		if err != nil {
			return err
		}
		if packageContentDiffer(packageFromTenant, packageFromGit) {
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
	// Clean up working directory
	err = os.RemoveAll(workDir + "/from_tenant")
	if err != nil {
		return err
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

		gitArtifactPath := fmt.Sprintf("%v/%v", gitSrcDir, directoryName)
		if file.Exists(fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitArtifactPath)) {
			// (1) If artifact already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")

			// Diff artifact contents
			dirDiffer, err := dt.CompareContent(downloadedArtifactPath, gitArtifactPath, scriptCollectionMap, "remote")
			if err != nil {
				return err
			}

			if dirDiffer {
				log.Info().Msg("🏆 Changes detected and will be updated to Git")
				// Update the changes into the Git directory
				err = dt.CopyContent(downloadedArtifactPath, gitArtifactPath)
				if err != nil {
					return err
				}
			} else {
				log.Info().Msg("🏆 No changes detected. Update to Git not required")
			}

		} else { // (2) If artifact does not exist in Git, then add it
			log.Info().Msgf("🏆 Artifact %v does not exist, and will be added to Git", artifact.Id)
			// Update the script collection in IFlow BPMN2 XML before syncing to Git
			if artifact.ArtifactType == "Integration" {
				err = file.UpdateBPMN(downloadedArtifactPath, scriptCollectionMap)
				if err != nil {
					return err
				}
			}
			err = file.ReplaceDir(downloadedArtifactPath, gitArtifactPath)
			if err != nil {
				return err
			}
		}
	}

	// Clean up working directory
	err = os.RemoveAll(workDir + "/download")
	if err != nil {
		return err
	}

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

func packageContentDiffer(source *odata.PackageSingleData, target *odata.PackageSingleData) bool {
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
