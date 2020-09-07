package file

import (
	"co-pilot/pkg/logger"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var log = logger.Context()

func Find(fileSuffix string, dir string) (result string, err error) {
	err = filepath.Walk(dir,
		func(path string, fi os.FileInfo, errIn error) error {
			if strings.HasSuffix(path, fileSuffix) {
				result = path
				return io.EOF
			}
			return nil
		})

	if err == io.EOF {
		err = nil
	}
	return
}

func ReadJson(file string, parsed interface{}) error {
	byteValue, err := Open(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &parsed)
	if err != nil {
		return err
	}

	return nil
}

func ReadXml(file string, parsed interface{}) error {
	byteValue, err := Open(file)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(byteValue, &parsed)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Open(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}

	defer file.Close()

	return byteValue, nil
}

func Overwrite(lines []string, filePath string) error {
	return ioutil.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func Copy(sourceFile string, destinationFile string) error {
	if Exists(destinationFile) {
		log.Warnf("%s already exists", destinationFile)
		return nil
	}

	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	destinationParts := strings.Split(destinationFile, "/")
	destinationDir := strings.Join(destinationParts[:len(destinationParts)-1], "/")
	if !Exists(destinationDir) {
		err = CreateDirectory(destinationDir)
		if err != nil {
			return err
		}
	}

	log.Infof("copying %s", sourceFile)
	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func RelPath(sourceDirectory string, filePath string) (string, error) {

	directoryParts := strings.Split(sourceDirectory, "/")
	fileParts := strings.Split(filePath, "/")

	if len(directoryParts) >= len(fileParts) {
		return "", errors.New("directory cannot be deeper than filePath")
	}

	cut := 0

	for i := range directoryParts {
		if directoryParts[i] == fileParts[i] {
			cut += 1
		} else {
			break
		}
	}

	return strings.Join(fileParts[cut:], "/"), nil
}

func CreateDirectory(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return err
		}
	}

	return nil
}
