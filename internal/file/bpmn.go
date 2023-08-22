package file

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/rs/zerolog/log"
	"os"
)

func UpdateBPMN(artifactDir string, scriptMap string) error {
	if scriptMap != "" {
		// Extract collection into key pairs
		output := map[string]string{}
		log.Debug().Msgf("Updating files in %v with collection %v", artifactDir, scriptMap)
		pairs := str.ExtractDelimitedValues(scriptMap, ",")
		for _, pair := range pairs {
			srcTgt := str.ExtractDelimitedValues(pair, "=")
			output[srcTgt[0]] = srcTgt[1]
		}

		if len(output) != 0 {
			bpmnDir := fmt.Sprintf("%v/src/main/resources/scenarioflows/integrationflow", artifactDir)
			entries, err := os.ReadDir(bpmnDir)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					artifactFile := fmt.Sprintf("%v/%v", bpmnDir, entry.Name())

					err = updateXML(artifactFile, output)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func updateXML(filePath string, scripts map[string]string) error {
	log.Info().Msgf("Processing BPMN2 file %v", filePath)
	// Read XML file into tree
	doc := etree.NewDocument()
	err := doc.ReadFromFile(filePath)
	if err != nil {
		return err
	}

	contentUpdated := false
	// Look for occurrence of scriptBundleId
	for _, bundles := range doc.FindElements("//ifl:property[key='scriptBundleId']") {
		v := bundles.SelectElement("value")
		sourceValue := v.Text()
		targetValue := scripts[sourceValue]
		if sourceValue != "" && targetValue != "" {
			log.Debug().Msgf("Changing scriptBundleId from %v to %v", sourceValue, targetValue)
			v.SetText(targetValue)
			contentUpdated = true
		}
	}
	// Update the BPMN XML file with the changes
	if contentUpdated {
		err = doc.WriteToFile(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}
