# Co-pilot
A little "go help" for the Java/Kotlin developers using Maven.

Current main capability? 
Upgrade your pom.xml dependencies to latest and greatest! 

Why?
- No installs of maven-plugins required, so if you a working in a multi-repo developer environment with lots of 2party dependencies and repos, you can easily upgrade them with `co-pilot upgrade 2party`. 
- Brings natural semantics and support for different types of dependencies to the table: Kotlin, 2party, spring-boot (curated dependencies), (other) 3party   
- Can be used as a library for other go-projects automating the upgrade process
- Easy and fast
- Brings feature to the table, not found anywhere else, stay tuned!

Heads up!
- co-pilot rewrites your pom.xml, so make sure you have your pom.xml committed before testing out co-pilot
- start with `co-pilot format pom`, verify that the rewrite of the pom.xml is ok, commit, and from now on you will easily see the diff that co-pilot introduces with ```co-pilot upgrade <2party|3party|spring-boot|plugins|all>```
- or just use  `co-pilot status` (no rewrite) and manually upgrade your pom.xml based on what is reported as outdated, current option if you need to keep your pom.xml formatting
  
Requirement: https://golang.org/doc/install

```shell script
   _____                  _ _       _
  / ____|                (_) |     | |
 | |     ___ ______ _ __  _| | ___ | |_
 | |    / _ \______| '_ \| | |/ _ \| __|
 | |___| (_) |     | |_) | | | (_) | |_
  \_____\___/      | .__/|_|_|\___/ \__|
                   | |
                   |_|
== version: v0.2.10, built: 2020-09-14 14:03 ==
Co-pilot is a developer tool for automating common tasks on a spring boot project

Usage:
  co-pilot [command]

Available Commands:
  analyze     analyze options
  bitbucket   Bitbucket functionality
  config      Config settings for co-pilot
  deprecated  Deprecated detection and patching functionalities for projects
  format      Format functionality for a project
  help        Help about any command
  maven       Maven options
  merge       Merge functionalities for files to a project
  project     Project options
  spring      Spring boot tools
  status      Prints project status
  upgrade     Upgrade options

Flags:
      --debug   turn on debug output
  -h, --help    help for co-pilot

Use "co-pilot [command] --help" for more information about a command.

```

## Install
```shell script
make
```

## Help
```shell script
co-pilot
```

