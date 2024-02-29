package main

import (
	"bufio"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

//go:embed A.txt template.html favicon.ico
var F embed.FS

type Question struct {
	ID       string
	Num      string // 题号
	Question string
	Options  map[string]string
}

func ReadQuestionsFromFile(filePath string) ([]Question, error) {
	file, err := F.Open(filePath)
	// file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var questions []Question
	scanner := bufio.NewScanner(file)
	var q Question
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[I]") {
			//第一次扫描，ID为空
			if q.ID != "" {
				questions = append(questions, q)
			}
			q = Question{Options: make(map[string]string)}
			q.ID = strings.TrimPrefix(line, "[I]")
		} else if strings.HasPrefix(line, "[Q]") {
			q.Question = strings.TrimPrefix(line, "[Q]")
		} else if strings.HasPrefix(line, "[A]") {
			//以"]"分割一个选项,后面的内容作为map value,前面的部分"[A"使用切片截取[1:0]得到'A',作为key
			parts := strings.SplitN(line, "]", 2)
			if len(parts) == 2 {
				q.Options[parts[0][1:]] = parts[1]
			}
		}
	}
	if q.ID != "" { // 添加最后一个题目
		questions = append(questions, q)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return questions, nil
}

func assignQuestionNumbers(questions []Question) {
	for i, _ := range questions {
		questions[i].Num = fmt.Sprintf("%03d", i+1)
	}
}

func renderQuestions(w http.ResponseWriter, questions []Question) {
	//tmpl := template.Must(template.New("").ParseFS(F, "template.html"))
	tmpl, _ := template.ParseFS(F, "template.html")
	tmpl.Execute(w, questions)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		questions, err := ReadQuestionsFromFile("A.txt")
		if err != nil {
			http.Error(w, "Failed to read questions", http.StatusInternalServerError)
			return
		}

		// 对questions切片按ID排序
		sort.Slice(questions, func(i, j int) bool {
			return questions[i].ID < questions[j].ID
		})

		// 分配题号
		assignQuestionNumbers(questions)

		renderQuestions(w, questions)
	})

	http.Handle("/favicon.ico", http.FileServer(http.FS(F)))
	http.Handle("/font.ttf", http.FileServer(http.FS(F)))
	log.Print("监听0.0.0.0:5488")
	http.ListenAndServe(":5488", nil)
	http.ListenAndServe("[::1]:5488", nil)

}