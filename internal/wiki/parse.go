package wiki

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"unicode/utf8"
)

/*
extract data

case 1:
	학습날짜 2020-10-02 : datetime
	학습시간 24시간제(위치) : string to int
	학습범위 및 주제 : string
	동료학습 방법 : string(x or nil)
case 2:
	학습날짜 2021년 1월 22일 : string
	학습시간 : 오전 10시 ~ 2시
	학습범위 및 주제 : string
	동료학습 방법 : 인터넷 검색

최종:
	학습날짜 : report 파일 이름에서 추출
	학습시간 : 24시간제
	학습범위 및 주제 : string 통으로
	동료학습 방법 : intraID
*/
// https://mholt.github.io/json-to-go/
// json to go struct

// 1day
type ReportInfo struct {
	Year       int      `json:"date"`  // yyyy
	Month      int      `json:"Month"` // mm
	Date       int      `json:"date"`  // dd
	Day        int      `json:"day"`   // Mon: 1, Tue: 2 ... Sun: 7
	StudyTime  int      `json:"studyTime"`
	StudyTheme string   `json:"studyTheme"`
	Cadet      []string `json:"cadets"`
}

func (repo *ReportInfo) ParseDate(filename string) {
	m := regexp.MustCompile(`[.,\- \(\)]`)
	filteredValue := m.ReplaceAll([]byte(filename), []byte(""))

	repo.Year, _ = strconv.Atoi(string(filteredValue[:4]))
	repo.Month, _ = strconv.Atoi(string(filteredValue[4:6]))
	repo.Date, _ = strconv.Atoi(string(filteredValue[6:8]))
	day, _ := utf8.DecodeRune(filteredValue[8:11])

	switch {
	case day == '월':
		repo.Day = 1
		break
	case day == '화':
		repo.Day = 2
		break
	case day == '수':
		repo.Day = 3
		break
	case day == '목':
		repo.Day = 4
		break
	case day == '금':
		repo.Day = 5
		break
	case day == '토':
		repo.Day = 6
		break
	case day == '일':
		repo.Day = 7
		break
	}
}

func DecodeFileName(filename string) string {
	decodedFileName, _ := url.QueryUnescape(filename)
	return decodedFileName
}

func StudyTimeStamp2Minute(studyTime []byte) int {
	timeStampMatcher := regexp.MustCompile(`\d\d?:\d\d?`)
	studyTimeStamp := timeStampMatcher.FindAll(studyTime, 2)

	timeMatcher := regexp.MustCompile(`\d\d?`)
	startTime := timeMatcher.FindAll(studyTimeStamp[0], 2)
	endTime := timeMatcher.FindAll(studyTimeStamp[1], 2)

	var startHour, startMinute, endHour, endMinute int
	if startTime[0][0] == '0' {
		startHour, _ = strconv.Atoi(string(startTime[0][1]))
	} else {
		startHour, _ = strconv.Atoi(string(startTime[0]))
	}
	if startTime[1][0] == '0' {
		startMinute, _ = strconv.Atoi(string(startTime[1][1]))
	} else {
		startMinute, _ = strconv.Atoi(string(startTime[1]))
	}
	if endTime[0][0] == '0' {
		endHour, _ = strconv.Atoi(string(endTime[0][1]))
	} else {
		endHour, _ = strconv.Atoi(string(endTime[0]))
	}
	if endTime[1][0] == '0' {
		endMinute, _ = strconv.Atoi(string(endTime[1][1]))
	} else {
		endMinute, _ = strconv.Atoi(string(endTime[1]))
	}

	log.Printf("\n\tstartHour: %d\n\tstartMinute: %d\n\tendHour: %d\n\tendMinute: %d\n", startHour, startMinute, endHour, endMinute)

	totalStudyTime := (endHour*60 + endMinute) - (startHour*60 - startMinute)
	log.Printf("\n\t totalStudyTime: %d\n", totalStudyTime)
	return totalStudyTime
}

func (repo *ReportInfo) ParseReport(filename string) {
	// Study Time
	raw_data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	// 24시 hh:mm ~ hh:mm
	m := regexp.MustCompile(`\d\d?:\d\d? ?(-|~) ?\d\d?:?\d?\d?`)
	data := m.FindAll(raw_data, -1)
	var studyTime []string
	for _, d := range data {
		log.Printf("\n\t raw time: %q", d)
		StudyTimeStamp2Minute(d)
		studyTime = append(studyTime, string(d))
	}

}

func GetReportInfo(filename string) *ReportInfo {
	info, _ := os.Stat(filename)
	reportInfo := &ReportInfo{}
	reportInfo.ParseDate(DecodeFileName(info.Name()))
	reportInfo.ParseReport(filename)
	return nil
}

func GetReport(intraID string) { //*ReportInfo {
	//	files, err := filepath.Glob(wikiRepoPath + intraID + "/20[0-9]{2}[.-, ]?[0-9]{2}[.-, ]?[0-9]{2}*.md")
	files, err := filepath.Glob(wikiRepoPath + intraID + "/*.md")

	log.Println("Get report")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	//	var reportInfo []ReportInfo
	for i := range files {
		//log.Printf("file: %v\n", files[i])
		GetReportInfo(files[i])
	}

}

/*
func ParseDate(intraID string) bool {
	if _, err := os.Stat(wikiRepoPath + intraID); os.IsNotExist(err) {
		log.Printf("Not exist [%v] repo", intraID)
		return false
	}

	return true
}
*/

/* 학습시간
24시형
\d\d?:\d\d? ?(-|~) ?\d\d?:?\d?\d?

am/pm형
// \d\d?(am|pm).*\d\d?

시간형
\d\d? ?시간
*/