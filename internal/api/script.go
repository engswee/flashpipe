package api

import (
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
	"os"
)

type ScriptCollection struct {
	exe *httpclnt.HTTPExecuter
	typ string
}

// NewScriptCollection returns an initialised ScriptCollection instance.
func NewScriptCollection(exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	sc := new(ScriptCollection)
	sc.exe = exe
	sc.typ = "ScriptCollection"
	return sc
}

func (sc *ScriptCollection) Create(id string, name string, packageId string, artifactDir string) error {
	return create(id, name, packageId, artifactDir, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Update(id string, name string, packageId string, artifactDir string) (err error) {
	return update(id, name, packageId, artifactDir, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Deploy(id string) (err error) {
	return deploy(id, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Delete(id string) (err error) {
	return deleteCall(id, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Get(id string, version string) (string, string, bool, error) {
	return get(id, version, sc.typ, sc.exe)
}
func (sc *ScriptCollection) Download(targetFile string, id string) error {
	return download(targetFile, id, sc.typ, sc.exe)
}
func (sc *ScriptCollection) CopyContent(srcDir string, tgtDir string) error {
	// Copy META-INF and /src/main/resources separately so that other directories like QA, STG, PRD not copied
	err := file.ReplaceDir(srcDir+"/META-INF", tgtDir+"/META-INF")
	if err != nil {
		return err
	}

	// It is technically possible to have an empty script collection
	if file.Exists(srcDir + "/src/main/resources") {
		// As long as it exists in source, will copy/replace it in target
		err = file.ReplaceDir(srcDir+"/src/main/resources", tgtDir+"/src/main/resources")
		if err != nil {
			return err
		}
	} else if file.Exists(tgtDir + "/src/main/resources") {
		// If resources does not exist in source but exists in target, then remove target
		err = os.RemoveAll(tgtDir + "/src/main/resources")
		if err != nil {
			return err
		}
	}
	// Copy also metainfo.prop that contains the description if it is available
	if file.Exists(srcDir + "/metainfo.prop") {
		err = file.CopyFile(srcDir+"/metainfo.prop", tgtDir+"/metainfo.prop")
		if err != nil {
			return err
		}
	}
	return nil
}
func (sc *ScriptCollection) CompareContent(srcDir string, tgtDir string, _ []string, _ string) (bool, error) {
	// Diff directories
	log.Info().Msg("Checking for changes in META-INF directory")
	metaDiffer := file.DiffDirectories(srcDir+"/META-INF", tgtDir+"/META-INF")
	// It is technically possible to have an empty script collection
	if file.Exists(srcDir+"/src/main/resources") && file.Exists(tgtDir+"/src/main/resources") {
		return metaDiffer || diffContent(srcDir, tgtDir), nil
	} else if !file.Exists(srcDir+"/src/main/resources") && !file.Exists(tgtDir+"/src/main/resources") {
		log.Warn().Msg("Skipping diff as /src/main/resources does not exist in both source and target")
		log.Info().Msg("Checking for changes in metainfo.prop")
		metainfoDiffer := DiffOptionalFile(srcDir, tgtDir, "metainfo.prop")
		return metaDiffer || metainfoDiffer, nil
	}
	log.Info().Msg("Directory /src/main/resources does not exist in either source or target")
	return true, nil
}
