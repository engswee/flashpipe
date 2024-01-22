package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/go-errors/errors"
	"github.com/magiconair/properties"
	"github.com/rs/zerolog/log"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Synchroniser struct {
	exe *httpclnt.HTTPExecuter
	ip  *api.IntegrationPackage
}

func New(exe *httpclnt.HTTPExecuter) *Synchroniser {
	s := new(Synchroniser)
	s.exe = exe
	s.ip = api.NewIntegrationPackage(exe)
	return s
}

func (s *Synchroniser) PackageToGit(packageDataFromTenant *api.PackageSingleData, packageId string, workDir string, artifactsDir string) error {
	// Create temp directory in working dir
	err := os.MkdirAll(workDir+"/from_tenant", os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	log.Info().Msg("Storing package details from tenant for comparison")
	// Write package details from tenant to file
	tenantFile := fmt.Sprintf("%v/from_tenant/%v.json", workDir, packageId)
	f, err := os.Create(tenantFile)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	defer f.Close()
	content, err := json.MarshalIndent(packageDataFromTenant, "", "  ")
	if err != nil {
		return errors.Wrap(err, 0)
	}
	_, err = f.Write(content)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Get existing package details file if it exists and compare values
	gitSourceFile := fmt.Sprintf("%v/%v.json", artifactsDir, packageId)
	if file.Exists(gitSourceFile) {
		packageDataFromGit, err := api.GetPackageDetails(tenantFile)
		if err != nil {
			return err
		}
		if packageContentDiffer(packageDataFromTenant, packageDataFromGit) {
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
		return errors.Wrap(err, 0)
	}

	return nil
}

func (s *Synchroniser) VerifyDownloadablePackage(packageId string) (packageDataFromTenant *api.PackageSingleData, readOnly bool, packageExists bool, err error) {
	// Verify the package is downloadable (not read only)
	packageDataFromTenant, readOnly, packageExists, err = s.ip.Get(packageId)
	if err != nil {
		return nil, false, false, err
	}
	if !packageExists {
		return nil, false, false, fmt.Errorf("Package %v does not exist", packageId)
	}
	if readOnly {
		log.Warn().Msgf("Skipping package %v as it is Configure-only and cannot be downloaded", packageId)
	}
	return
}

func (s *Synchroniser) ArtifactsToGit(packageId string, workDir string, artifactsDir string, includedIds []string, excludedIds []string, draftHandling string, dirNamingType string, scriptCollectionMap []string) error {
	// Get all design time artifacts of package
	log.Info().Msgf("Getting artifacts in integration package %v", packageId)
	artifacts, err := s.ip.GetAllArtifacts(packageId)
	if err != nil {
		return err
	}

	// Create temp directories in working dir
	err = os.MkdirAll(workDir+"/download", os.ModePerm)
	if err != nil {
		return errors.Wrap(err, 0)
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
		dt := api.NewDesigntimeArtifact(artifact.ArtifactType, s.exe)
		targetDownloadFile := fmt.Sprintf("%v/download/%v.zip", workDir, artifact.Id)
		err = dt.Download(targetDownloadFile, artifact.Id)
		if err != nil {
			return err
		}

		// TODO - override directory name using key value pair - to cater for syncing artifact from different environment
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

		gitArtifactPath := fmt.Sprintf("%v/%v", artifactsDir, directoryName)
		if file.Exists(fmt.Sprintf("%v/META-INF/MANIFEST.MF", gitArtifactPath)) {
			// (1) If artifact already exists in Git, then compare and update
			log.Info().Msg("Comparing content from tenant against Git")

			// Diff artifact contents
			dirDiffer, err := dt.CompareContent(downloadedArtifactPath, gitArtifactPath, scriptCollectionMap, "git")
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
		return errors.Wrap(err, 0)
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msgf("🏆 Completed processing of artifacts in integration package %v", packageId)
	return nil
}

func filterArtifacts(artifacts []*api.ArtifactDetails, includedIds []string, excludedIds []string) ([]*api.ArtifactDetails, error) {
	var output []*api.ArtifactDetails
	// Trim whitespace from IDs
	includedIds = str.TrimSlice(includedIds)
	excludedIds = str.TrimSlice(excludedIds)
	if len(includedIds) > 0 {
		for _, id := range includedIds {
			artifact := api.FindArtifactById(id, artifacts)
			if artifact != nil {
				output = append(output, artifact)
			} else {
				return nil, fmt.Errorf("Artifact %v in --ids-include does not exist", id)
			}
		}
		return output, nil
	} else if len(excludedIds) > 0 {
		for _, id := range excludedIds {
			artifact := api.FindArtifactById(id, artifacts)
			if artifact == nil {
				return nil, fmt.Errorf("Artifact %v in --ids-exclude does not exist", id)
			}
		}
		for _, artifact := range artifacts {
			if !slices.Contains(excludedIds, artifact.Id) {
				output = append(output, artifact)
			}
		}
		return output, nil
	}
	return artifacts, nil
}

func packageContentDiffer(source *api.PackageSingleData, target *api.PackageSingleData) bool {
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

func (s *Synchroniser) ArtifactsToTenant(packageId string, workDir string, artifactsDir string, includedIds []string, excludedIds []string) error {
	// Get directory list
	baseSourceDir := filepath.Clean(artifactsDir)
	entries, err := os.ReadDir(baseSourceDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	artifactDirFound := false
	for _, entry := range entries {
		manifestPath := fmt.Sprintf("%v/%v/META-INF/MANIFEST.MF", baseSourceDir, entry.Name())
		if entry.IsDir() && file.Exists(manifestPath) {
			artifactDirFound = true
			artifactDir := fmt.Sprintf("%v/%v", baseSourceDir, entry.Name())
			log.Info().Msg("---------------------------------------------------------------------------------")
			log.Info().Msgf("Processing directory %v", artifactDir)
			paramFile := fmt.Sprintf("%v/src/main/resouces/parameters/prop", artifactDir)

			headers, err := GetManifestHeaders(manifestPath)
			if err != nil {
				return err
			}

			artifactId := headers.Get("Bundle-SymbolicName")
			// remove spaces then remove ;singleton:=true
			artifactId = strings.ReplaceAll(artifactId, " ", "")
			artifactId = strings.ReplaceAll(artifactId, ";singleton:=true", "")

			// Filter in/out artifacts
			if len(includedIds) > 0 {
				if !slices.Contains(includedIds, artifactId) {
					log.Warn().Msgf("Skipping artifact %v as it is not in --ids-include", artifactId)
					continue
				}
			}
			if len(excludedIds) > 0 {
				if slices.Contains(excludedIds, artifactId) {
					log.Warn().Msgf("Skipping artifact %v as it is in --ids-exclude", artifactId)
					continue
				}
			}

			artifactName := headers.Get("Bundle-Name")
			// remove spaces due to length of bundle name exceeding MANIFEST.MF width
			artifactName = strings.ReplaceAll(artifactName, " ", "")
			artifactType := headers.Get("SAP-BundleType")
			if artifactType == "IntegrationFlow" {
				artifactType = "Integration"
			}

			log.Info().Msgf("📢 Begin processing for artifact %v", artifactId)
			err = s.SingleArtifactToTenant(artifactId, artifactName, artifactType, packageId, artifactDir, workDir, paramFile, nil)
			if err != nil {
				return err
			}
		}
	}
	if !artifactDirFound {
		log.Warn().Msgf("No directory with artifact contents found in %v", baseSourceDir)
	}
	return nil
}

func GetManifestHeaders(manifestPath string) (textproto.MIMEHeader, error) {
	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer manifestFile.Close()

	tp := textproto.NewReader(bufio.NewReader(manifestFile))
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	return headers, nil
}

func (s *Synchroniser) SingleArtifactToTenant(artifactId, artifactName, artifactType, packageId, artifactDir, workDir, parametersFile string, scriptMap []string) error {
	dt := api.NewDesigntimeArtifact(artifactType, s.exe)

	exists, err := artifactExists(artifactId, artifactType, packageId, dt, s.ip)
	if err != nil {
		return err
	}

	if !exists {
		log.Info().Msgf("Artifact %v will be created", artifactId)
		if artifactType == "Integration" {
			err = file.UpdateBPMN(artifactDir, scriptMap)
			if err != nil {
				return err
			}
		}

		err = prepareUploadDir(workDir, artifactDir, dt)
		if err != nil {
			return err
		}

		err = createArtifact(artifactId, artifactName, packageId, workDir+"/upload", dt)
		if err != nil {
			return err
		}

		log.Info().Msg("🏆 Designtime artifact created successfully")
	} else {
		log.Info().Msg("Checking if designtime artifact needs to be updated")

		zipFile := fmt.Sprintf("%v/%v.zip", workDir, artifactId)
		err = dt.Download(zipFile, artifactId)
		if err != nil {
			return err
		}

		changesFound, err := compareArtifactContents(workDir, zipFile, artifactDir, scriptMap, dt)
		if err != nil {
			return err
		}

		if changesFound == true {
			log.Info().Msg("Changes found in designtime artifact. Designtime artifact will be updated in CPI tenant")
			err = prepareUploadDir(workDir, artifactDir, dt)
			if err != nil {
				return err
			}
			err = updateArtifact(artifactId, artifactName, packageId, workDir+"/upload", dt)
			if err != nil {
				return err
			}

			designtimeVersion, _, err := dt.Get(artifactId, "active")
			if err != nil {
				return err
			}
			r := api.NewRuntime(s.exe)
			runtimeVersion, _, err := r.Get(artifactId)
			if err != nil {
				return err
			}
			if runtimeVersion == designtimeVersion {
				log.Info().Msg("Undeploying existing runtime artifact with same version number due to changes in design")
				err = r.UnDeploy(artifactId)
				if err != nil {
					return err
				}
			}

			log.Info().Msg("🏆 Designtime artifact updated successfully")
		} else {
			log.Info().Msg("🏆 No changes detected. Designtime artifact does not need to be updated")
		}

		if artifactType == "Integration" && file.Exists(parametersFile) {
			log.Info().Msg("Updating configured parameter(s) of Integration designtime artifact where necessary")
			err = updateConfiguration(artifactId, parametersFile, s.exe)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func artifactExists(artifactId string, artifactType string, packageId string, dt api.DesigntimeArtifact, ip *api.IntegrationPackage) (bool, error) {
	_, exists, err := dt.Get(artifactId, "active")
	if err != nil {
		return false, err
	}
	if exists {
		log.Info().Msgf("Active version of artifact %v exists", artifactId)
		//  Check if version is in draft mode
		var details []*api.ArtifactDetails
		details, err = ip.GetArtifactsData(packageId, artifactType)
		if err != nil {
			return false, err
		}
		artifact := api.FindArtifactById(artifactId, details)
		if artifact == nil {
			return false, fmt.Errorf("Artifact %v not found in package %v", artifactId, packageId)
		}
		if artifact.IsDraft {
			return false, fmt.Errorf("Artifact %v is in Draft state. Save Version of artifact in Web UI first!", artifactId)
		}
		return true, nil
	} else {
		log.Info().Msgf("Active version of artifact %v does not exist", artifactId)
		return false, nil
	}
}

func prepareUploadDir(workDir string, artifactDir string, dt api.DesigntimeArtifact) error {
	// Clean up previous uploads
	uploadDir := workDir + "/upload"
	err := os.RemoveAll(uploadDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return dt.CopyContent(artifactDir, uploadDir)
}

func createArtifact(artifactId string, artifactName string, packageId string, artifactDir string, dt api.DesigntimeArtifact) error {
	err := dt.Create(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}

func updateArtifact(artifactId string, artifactName string, packageId string, artifactDir string, dt api.DesigntimeArtifact) error {
	err := dt.Update(artifactId, artifactName, packageId, artifactDir)
	if err != nil {
		return err
	}
	return nil
}

func compareArtifactContents(workDir string, zipFile string, artifactDir string, scriptMap []string, dt api.DesigntimeArtifact) (bool, error) {
	tgtDir := fmt.Sprintf("%v/download", workDir)
	err := os.RemoveAll(tgtDir)
	if err != nil {
		return false, errors.Wrap(err, 0)
	}

	log.Info().Msgf("Unzipping downloaded designtime artifact %v to %v/download", zipFile, workDir)
	err = file.UnzipSource(zipFile, tgtDir)
	if err != nil {
		return false, err
	}

	return dt.CompareContent(artifactDir, tgtDir, scriptMap, "tenant")
}

func updateConfiguration(artifactId string, parametersFile string, exe *httpclnt.HTTPExecuter) error {
	// Get configured parameters from tenant
	c := api.NewConfiguration(exe)
	tenantParameters, err := c.Get(artifactId, "active")
	if err != nil {
		return err
	}

	// Get parameters from parameters.prop file
	log.Info().Msgf("Getting parameters from %v file", parametersFile)
	fileParameters := properties.MustLoadFile(parametersFile, properties.UTF8)

	log.Info().Msg("Comparing parameters and updating where necessary")
	atLeastOneUpdated := false
	for _, result := range tenantParameters.Root.Results {
		if result.DataType != "custom:schedule" { // TODO - handle translation to Cron
			// Skip updating for schedulers which require translation to Cron values
			fileValue := fileParameters.GetString(result.ParameterKey, "")
			if fileValue != "" && fileValue != result.ParameterValue {
				log.Info().Msgf("Parameter %v to be updated from %v to %v", result.ParameterKey, result.ParameterValue, fileValue)
				err = c.Update(artifactId, "active", result.ParameterKey, fileValue)
				if err != nil {
					return err
				}
				atLeastOneUpdated = true
			}
		}
	}
	if atLeastOneUpdated {
		r := api.NewRuntime(exe)
		version, _, err := r.Get(artifactId)
		if err != nil {
			return err
		}
		if version == "NOT_DEPLOYED" {
			log.Info().Msg("🏆 No existing runtime artifact deployed")
		} else {
			log.Info().Msg("🏆 Undeploying existing runtime artifact due to changes in configured parameters")
			err = r.UnDeploy(artifactId)
			if err != nil {
				return err
			}
		}
	} else {
		log.Info().Msg("🏆 No updates required for configured parameters")
	}
	return nil
}
