package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/sqweek/dialog"
	xmlXmind "github.com/xiazemin/xmind-go"
	"io/ioutil"
	"log"
	"path/filepath"
)

const (
	contentFilename = "content.json"
)

var outFilename = ""

type (
	Topic struct {
		Id       string `json:"id"`
		Title    string `json:"title"`
		Children struct {
			Attached []Topic `json:"attached"`
		} `json:"children"`
	}

	Sheet struct {
		Id        string `json:"id"`
		Class     string `json:"class"`
		Title     string `json:"title"`
		RootTopic Topic  `json:"rootTopic"`
	}
)
type Node struct {
	NodeID       string `json:"node_id"`
	TopicContent string `json:"topic_content"`
	ParentID     string `json:"parent_id,omitempty"`
}

func ReadXmind(fullFilename string) ([]Node, error) {
	zipFp, zipErr := zip.OpenReader(fullFilename)
	if zipErr != nil {
		return nil, fmt.Errorf("failed to open zip file %s: %s", fullFilename, zipErr)
	}
	defer zipFp.Close()

	sheets := make([]Sheet, 0)
	for _, file := range zipFp.File {
		if file.FileInfo().Name() != contentFilename {
			continue
		}

		contentFp, openErr := file.Open()
		if openErr != nil {
			log.Fatalf("Failed to open file %s: %s\n", contentFilename, openErr)
		}

		contentByte, readErr := ioutil.ReadAll(contentFp)
		if readErr != nil {
			log.Fatalf("Failed to read file %s: %s\n", contentFilename, readErr)
		}

		jsonErr := json.Unmarshal(contentByte, &sheets)
		if jsonErr != nil {
			log.Fatalf("Failed to parse json to Sheet: %s\n", jsonErr)
		}
		fmt.Printf(string(contentByte))

		//prettyJson, _ := json.MarshalIndent(sheets, "", "  ")
		//log.Println(string(prettyJson))
	}
	// 遍历并拼接每个分支的子节点以及他的子节点，直到叶子节点
	concatenatedTitles := make(map[string]string)
	var nodes []Node
	concatenateBranches(sheets[0].RootTopic, "", concatenatedTitles)
	nodes = append(nodes, Node{
		NodeID:       "1",
		TopicContent: sheets[0].RootTopic.Title,
	})
	outFilename = sheets[0].RootTopic.Title
	for id, titles := range concatenatedTitles {
		//fmt.Printf("Branch %s titles: %s\n", id, titles)
		nodes = append(nodes, Node{
			NodeID:       id,
			TopicContent: titles,
			ParentID:     "1",
		})
	}
	fmt.Println(nodes)
	return nodes, nil

}
func concatenateBranches(topic Topic, prefix string, result map[string]string) {
	for _, child := range topic.Children.Attached {
		if prefix != "" {
			concatenateBranches(child, prefix+"-"+child.Title, result)
		} else {
			concatenateBranches(child, prefix+child.Title, result)
		}
	}
	if len(topic.Children.Attached) == 0 {
		result[topic.Id] = prefix
	}
}
func main() {

	filename, err := selectXMindFile()
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := ReadXmind(filename)
	jsonData, err := json.MarshalIndent(nodes, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	err = xmlXmind.SaveSheets(outFilename+"_output.xmind", string(jsonData))
	if err != nil {
		panic(err)
	}
}
func selectXMindFile() (string, error) {
	filename, err := dialog.File().Filter("XMind Files", "xmind").Load()
	if err != nil {
		return "", err
	}
	if filename == "" {
		return "", fmt.Errorf("no file selected")
	}
	ext := filepath.Ext(filename)
	if ext != ".xmind" {
		return "", fmt.Errorf("selected file is not an XMind file")
	}
	return filename, nil
}
