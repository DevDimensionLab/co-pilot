package template

import (
	"encoding/json"
	"fmt"
	"github.com/cbroglie/mustache"
	"github.com/devdimensionlab/co-pilot/pkg/config"
	"github.com/devdimensionlab/co-pilot/pkg/file"
	"github.com/devdimensionlab/co-pilot/pkg/maven"
	"github.com/devdimensionlab/co-pilot/pkg/mustache_render"
	"github.com/devdimensionlab/mvn-pom-mutator/pkg/pom"
	"os"
	"path/filepath"
	"strings"
)

const TemplatesDir = "templates"

type TemplateRoot struct {
	Templates []config.CloudTemplate `json:"templates"`
}

func SaveTemplateListMarkdown(gitCfg config.CloudConfig, markdownDocument string) (string, error) {
	readmePath := gitCfg.Implementation().Dir() + "/" + TemplatesDir + "/README.md"
	err := os.WriteFile(readmePath, []byte(markdownDocument), 0644)
	return readmePath, err
}

func TemplateListAsMarkdown(gitCfg config.CloudConfig, templates []config.CloudTemplate) (string, error) {
	templateMustacheModel, err := TemplateListToMustacheModel(templates)
	templateString, err := mustache_render.MarkdownMustacheTemplateString(gitCfg, "templates-markdown.mustache")
	if err != nil {
		return "", err
	}

	output, err := mustache.Render(templateString, templateMustacheModel)
	if err != nil {
		return "", err
	}

	return output, nil
}
func TemplateListToMustacheModel(templates []config.CloudTemplate) (map[string]interface{}, error) {
	root := TemplateRoot{
		Templates: templates,
	}
	jsonModel, err := json.MarshalIndent(root, "", "    ")
	if err != nil {
		return nil, err
	}
	//	println(string(jsonModel[:]))

	var dataModel map[string]interface{}
	err = json.Unmarshal(jsonModel, &dataModel)
	if err != nil {
		return nil, err
	}

	return dataModel, err
}

func MergeTemplates(cloudTemplates []config.CloudTemplate, targetProject config.Project) {
	for _, template := range cloudTemplates {
		log.Infof("applying Template %s", template.Name)
		if err := MergeTemplate(template, targetProject, false); err != nil {
			log.Warnf("%v", err)
		}
	}
}

func MergeTemplate(cloudTemplate config.CloudTemplate, targetProject config.Project, multiModuleCheck bool) error {
	if targetProject.IsDirtyGitRepo() {
		log.Warn(fmt.Sprintf("merging template %s into a dirty git repository %s", cloudTemplate.Name, targetProject.Path))
	} else {
		log.Info(fmt.Sprintf("merging template %s into %s", cloudTemplate.Name, targetProject.Path))
	}
	if err := merge(cloudTemplate.Project, targetProject, multiModuleCheck); err != nil {
		return err
	}

	return nil
}

func merge(sourceProject config.Project, targetProject config.Project, multiModuleCheck bool) error {
	sourceDir := sourceProject.Path
	files, err := filesToCopy(sourceDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		sourceRelPath, err := file.RelPath(sourceDir, f)
		if err != nil {
			return err
		}

		sourceRelPath = replacePathForSource(sourceRelPath, sourceProject.Config, targetProject.Config)
		targetPath := file.Path("%s/%s", targetProject.Path, sourceRelPath)

		if err = file.CopyOrMerge(f, targetPath); err != nil {
			return err
		}

		if err = file.SearchReplace(targetPath, sourceProject.Config.Package, targetProject.Config.Package); err != nil {
			return err
		}

		if strings.HasSuffix(targetPath, ".render") {
			if err := renderAndDelete(targetPath, targetProject.Config); err != nil {
				return err
			}
		}
	}

	// multimodule specific code -- fixes additional pom.xml files in cloud template, and cleans root src, etc
	if sourceProject.IsMultiModule() && multiModuleCheck {
		if err := mergeMultimodulePoms(targetProject); err != nil {
			return err
		}
		if err := cleanForMultiModule(targetProject); err != nil {
			return err
		}
	}

	if sourceProject.IsMavenProject() && targetProject.IsMavenProject() {
		return maven.MergeAndWritePomFiles(sourceProject, targetProject)
	}
	return nil
}

func mergeMultimodulePoms(targetProject config.Project) error {
	pomFiles, err := file.FindAll("pom.xml", []string{}, targetProject.Path)
	if err != nil {
		return err
	}
	for _, pomFile := range pomFiles {
		if pomFile == file.Path("%s/pom.xml", targetProject.Path) {
			continue
		}
		subModel, err := pom.GetModelFrom(pomFile)
		if err != nil {
			return err
		}
		subModel.Parent.GroupId = targetProject.Type.Model().GroupId
		subModel.Parent.ArtifactId = targetProject.Type.Model().ArtifactId
		subModel.Parent.Version = targetProject.Type.Model().Version
		if err := subModel.WriteToFile(pomFile, "    "); err != nil {
			return err
		}
	}

	return nil
}

func cleanForMultiModule(targetProject config.Project) error {
	// dependencies
	var dependencies []pom.Dependency
	for _, dep := range targetProject.Type.Model().Dependencies.Dependency {
		if dep.GroupId == "org.springframework.boot" {
			continue
		}
		dependencies = append(dependencies, dep)
	}
	targetProject.Type.Model().Dependencies.Dependency = dependencies

	// plugins
	var plugins []pom.Plugin
	for _, plug := range targetProject.Type.Model().Build.Plugins.Plugin {
		if plug.GroupId == "org.springframework.boot" {
			continue
		}
		plugins = append(plugins, plug)
	}
	targetProject.Type.Model().Build.Plugins.Plugin = plugins

	// src folder in root
	return file.DeleteAll(file.Path("%s/src", targetProject.Path))
}

func filesToCopy(sourceDir string) (files []string, err error) {
	ignores := getIgnores(sourceDir)
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		rootDir := sourceDir == strings.ReplaceAll(path, "/"+info.Name(), "")
		for _, ignore := range ignores {
			if (ignore == "pom.xml") && !rootDir {
				continue
			}
			if strings.Contains(path, ignore) {
				log.Debugf("ignoring %s in %s", path, ignores)
				return nil
			}
		}
		files = append(files, path)
		return nil
	})
	log.Debugf("filesToCopy: %s", files)
	return
}

func getIgnores(sourceDir string) (ignores []string) {
	gitIgnores, err := file.OpenIgnoreFile(file.Path("%s/.gitignore", sourceDir))
	if err != nil {
		log.Error(err)
	}
	ignores = append(ignores, gitIgnores...)

	coPilotIgnores, err := file.OpenIgnoreFile(file.Path("%s/.co-pilot.ignore", sourceDir))
	if err != nil {
		log.Error(err)
	}
	ignores = append(ignores, coPilotIgnores...)

	defaultIgnores := []string{
		"pom.xml",
		"co-pilot.json",
		"Application",
		".co-pilot.ignore",
		".gitignore",
		".mvn",
		"mvnw",
		"mvnw.cmd",
		".idea",
		".iml",
		".git",
	}
	ignores = append(ignores, defaultIgnores...)

	return
}

func replacePathForSource(sourceRelPath string, sourceConfig config.ProjectConfiguration, targetConfig config.ProjectConfiguration) string {
	var output = sourceRelPath

	if strings.Contains(output, ".kt") || strings.Contains(output, ".java") {
		if strings.Contains(output, "src/main") {
			output = strings.Replace(sourceRelPath, sourceConfig.SourceMainPath(), targetConfig.SourceMainPath(), 1)
		} else if strings.Contains(output, "src/test") {
			output = strings.Replace(sourceRelPath, sourceConfig.SourceTestPath(), targetConfig.SourceTestPath(), 1)
		}

		if output == sourceRelPath {
			log.Warnf("was not able to replace path for source file (.kt, .java), input and output path is the same %s", output)
		}
	}

	return output
}

func renderAndDelete(targetPath string, targetConfig interface{}) error {
	newTarget := strings.Replace(targetPath, ".render", "", 1)
	log.Infof("rendering %s into %s", targetPath, newTarget)
	if err := file.Render(targetPath, newTarget, targetConfig); err != nil {
		return err
	}

	log.Infof("deleting old render file %s", targetPath)
	return file.DeleteSingleFile(targetPath)
}
