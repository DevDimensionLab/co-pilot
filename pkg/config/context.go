package config

import (
	"fmt"
	"github.com/co-pilot-cli/co-pilot/pkg/file"
	"github.com/co-pilot-cli/co-pilot/pkg/logger"
)

type Context struct {
	Recursive       bool
	DryRun          bool
	TargetDirectory string
	DisableGit      bool
	ForceCloudSync  bool
	OpenInBrowser   bool
	Projects        []Project
	Err             error
	ProfilesPath    string
	LocalConfig     LocalConfig
	CloudConfig     CloudConfig
}

func (ctx *Context) FindAndPopulateMavenProjects() error {
	excludes := []string{
		"flattened-pom.xml",
		"/target/",
	}

	if ctx.Recursive {
		pomFiles, err := file.FindAll("pom.xml", excludes, ctx.TargetDirectory)
		if err != nil {
			return err
		}
		for _, pomFile := range pomFiles {
			project, err := InitProjectFromPomFile(pomFile)
			if err != nil {
				log.Warnln(err)
			}
			ctx.Projects = append(ctx.Projects, project)
		}
	} else {
		project, err := InitProjectFromDirectory(ctx.TargetDirectory)
		if err != nil {
			return err
		}
		ctx.Projects = append(ctx.Projects, project)
	}

	return nil
}

func (ctx Context) OnEachProject(description string, do ...func(project Project) error) {
	if ctx.Projects == nil || len(ctx.Projects) == 0 {
		log.Errorln("could not find any pom models in the context")
		return
	}

	for _, p := range ctx.Projects {
		if p.CloudConfig != nil {
			projectDefaults, err := p.CloudConfig.ProjectDefaults()
			if err != nil {
				log.Warnf("could not find a project-defaults.json file in cloud-config")
				log.Debugf("%v", err)
			}
			p.Config.Settings.mergeProjectDefaults(projectDefaults)
		}
		if p.Type == nil {
			log.Warnf("no project type defined for path: %s", p.Path)
			continue
		}

		log.Info(logger.White(fmt.Sprintf("%s in %s", description, p.Path)))

		if p.IsDirtyGitRepo() {
			log.Debugf("operating on a dirty git repo")
		}

		if do != nil {
			for _, job := range do {
				if job == nil {
					continue
				}
				err := job(p)
				if err != nil {
					log.Warnln(err)
					continue
				}
			}
		}

		if !ctx.DryRun {
			if err := p.SortAndWritePom(); err != nil {
				log.Warnln(err)
			}
		}
	}
}

func (ctx Context) OnRootProject(description string, do ...func(project Project) error) {
	if ctx.Projects == nil || len(ctx.Projects) == 0 {
		log.Errorln("could not find any pom models in the context")
		return
	}

	rootProject := ctx.Projects[0]
	if rootProject.Type == nil {
		log.Fatalln(fmt.Sprintf("no project type defined for path: %s", rootProject.Path))
	}
	log.Info(logger.White(fmt.Sprintf("%s for file %s", description, rootProject.Type.FilePath())))

	if rootProject.IsDirtyGitRepo() {
		log.Warnf("operating on a dirty git repo")
	}

	if do != nil {
		for _, job := range do {
			if job == nil {
				continue
			}
			err := job(rootProject)
			if err != nil {
				log.Warnln(err)
				continue
			}
		}
	}

	if !ctx.DryRun {
		if err := rootProject.SortAndWritePom(); err != nil {
			log.Warnln(err)
		}
	}
}

func (ctx *Context) LoadProfile(profilePath string) {
	ctx.LocalConfig = NewLocalConfig(profilePath)
	ctx.CloudConfig = OpenGitCloudConfig(profilePath)
	if !ctx.LocalConfig.Exists() {
		err := ctx.LocalConfig.TouchFile()
		if err != nil {
			log.Error(err)
		}
	}
}
