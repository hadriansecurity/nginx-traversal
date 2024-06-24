package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

func checkAliasWithoutTrailingSlash(filePath string) (bool, error) {
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return false, err
    }
    
    lines := strings.Split(string(content), "\n")
    inLocationBlock := false
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "location ") {
            inLocationBlock = true
            if !strings.HasSuffix(line, "/") {
                for _, innerLine := range lines {
                    innerLine = strings.TrimSpace(innerLine)
                    if strings.HasPrefix(innerLine, "alias ") {
                        return true, nil
                    }
                }
            }
        } else if inLocationBlock && strings.HasPrefix(line, "}") {
            inLocationBlock = false
        }
    }
    return false, nil
}

func findFilesWithAliasWithoutTrailingSlash(rootDir string) ([]string, error) {
    var result []string
    err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".conf") {
            hasError, err := checkAliasWithoutTrailingSlash(path)
            if err != nil {
                return err
            }
            if hasError {
                result = append(result, path)
            }
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return result, nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go /path/to/nginx/configs")
        return
    }
    rootDir := os.Args[1]
    filesWithAliasWithoutTrailingSlash, err := findFilesWithAliasWithoutTrailingSlash(rootDir)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    if len(filesWithAliasWithoutTrailingSlash) == 0 {
        fmt.Println("No files found with alias without trailing slash.")
        return
    }
    fmt.Println("Files with alias without trailing slash:")
    for _, file := range filesWithAliasWithoutTrailingSlash {
        fmt.Println(file)
    }
}

