package home

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gtk-attendance/model"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)
var FileName = "/data/5.xlsx"
var ConfName = ""

const (
	sheet1 = "员工刷卡记录表"
)

const (
	BeginRow = 4
	BeginLoop = 3

	GongHaoKey = 4
	GongHaoValue = 5
	XingMingKey = 10
	XingMingValue = 11

	// 考勤异常
	AttendanceAbnormal = -100
	AttendanceDeletion = -200
)

var (
  month string
  year string
)

func Calc(p *Home) error {

	rows, err := getRows()
	if err != nil{
		log.Fatal(err)
		return err
	}

	var userMap = make(map[string]map[string][]string)
	var key string
	for i, row := range rows {
		if i < BeginRow {
			if i == 2 {
				str :=  row[25]
				nyStr := strings.Split(strings.Split(str, "：")[1], "～")[0]
				year = strings.Split(nyStr, "-")[0]
				month = strings.Split(nyStr, "-")[1]
			}
			continue
		}
		if len(row) >10 && row[GongHaoKey] == "工号：" && row[XingMingKey] == "姓名：" {
			key = fmt.Sprintf("%s-%s", row[XingMingValue],row[GongHaoValue])
			userMap[key] = make(map[string][]string)
		} else if len(row) >1 && row[1] == "1" {
			for j:=1;j<len(row);j++{
				userMap[key][row[j]] = []string{}
			}
		} else {
			for j:=1;j<len(row);j++{
				times := strings.Split(row[j], "\n")
				for _, v := range times{
					if strings.TrimSpace(v) != "" {
						userMap[key][fmt.Sprintf("%d", j)] = append(userMap[key][fmt.Sprintf("%d", j)], v)
					}
				}
			}
		}
	}



	list := getUserList(userMap)
	sort.Sort(list)
	for i, v := range list{
		list[i].Overtime = calcOvertime(v)
	}

	path, err := createExcel(list, p.Folder())
	if err != nil{
		log.Fatal(err)
		return err
	}
	p.SetFile(path)
	return err
}

func getRows() ([][]string, error) {
	//dir, err := os.Getwd()
	//path := dir+ FileName
	f, err := excelize.OpenFile(FileName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	rows, err := f.GetRows(sheet1)
	if err != nil {
		log.Fatal(err)
		return rows, err
	}

	return rows, err
}

func getUserList(m map[string]map[string][]string)(list model.UserList) {
	for k, v := range m {
		var user model.User
		user.Name=strings.Split(k,"-")[0]
		user.Number=strings.Split(k,"-")[1]
		att := make([][]string, len(v)+1)
		for day, attendance := range v{
			index,err := strconv.Atoi(day)
			if err != nil{
				log.Fatal(err)
			}
			att[index] = attendance
			user.Attendance = att
		}
		list = append(list, user)
	}
	return list
}

func calcOvertime(user model.User) []float64  {
	attendance := user.Attendance
	overtime := make([]float64, len(attendance))
	for day, v := range attendance {
		if day < 1{
			continue
		}
		res, err := checkHoliday(day)
		if err != nil{
			log.Fatal(err)
		}

		var next []string
		if day +1 < len(attendance) {
			next = attendance[day+1]
		}
		dayStr := fmt.Sprintf("%02d", day)


		// 周末国假加班，最晚时间-8:30-1.5h
		if res {
			endStr := dayStr
			if len(v) == 0{
				fmt.Println(fmt.Sprintf("%s %s %d %s",user.Name, dayStr, 0, "缺勤"))
				overtime[day] = AttendanceDeletion
				continue
			} else if len(v) == 1 && len(next) == 3 {
				// 今天打卡一次，明天打卡3次（加班到凌晨导致）
				fmt.Println(fmt.Sprintf("%s %s %s",user.Name, dayStr, "加班到凌晨"))
				attendance[day] = append(attendance[day], next[0])
				attendance[day+1] = []string{next[1], next[2]}
				endStr = fmt.Sprintf("%02d", day+1)

			} else if len(v) == 1 && len(next) != 3 {
				fmt.Println(fmt.Sprintf("%s %s %s",user.Name, dayStr, "考勤异常"))
				overtime[day] = AttendanceAbnormal
				continue
			}
			begin := attendance[day][0]
			end := attendance[day][len(attendance[day])-1]
			sbTime := getHolidayShangBan(fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, begin), dayStr)
			time := fmt.Sprintf("%s-%s-%s %s", year, month, endStr, end)
			etime := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "18:20")
			timeA := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "17:30")

			ztime1 := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "12:00")
			ztime2 := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "13:00")

			// 中午吃饭时间打卡，按照13:00上班
			if Time(sbTime).After(Time(ztime1)) && Time(sbTime).Before(Time(ztime2)){
				sbTime = ztime2
			}

			// 一定是12:00之前上班的
			if Time(sbTime).Sub(Time(ztime2)) < 0 {
				// 18:20后加班，减去1.5;18:20前加班，减去1;
				if  Time(time).After(Time(etime)){
					subM := Time(time).Sub(Time(sbTime))
					fmt.Println(fmt.Sprintf("%s %s %f %s",user.Name, dayStr, Binary(subM.Hours())-1.5, "小时"))
					overtime[day] = Binary(subM.Hours())-1.5

				} else {
					subM := Time(timeA).Sub(Time(sbTime))
					fmt.Println(fmt.Sprintf("%s %s %f %s",user.Name, dayStr, Binary(subM.Hours())-1, "小时"))
					overtime[day] = Binary(subM.Hours())-1
				}

			} else {
				// 一定是13:00之后上班的

				// 18:20后加班，减去0.5;18:20前加班，减去0;
				if  Time(time).After(Time(etime)){
					subM := Time(time).Sub(Time(sbTime))
					fmt.Println(fmt.Sprintf("%s %s %f %s",user.Name, dayStr, Binary(subM.Hours())-1.5, "小时"))
					overtime[day] = Binary(subM.Hours())-0.5

				} else {
					subM := Time(timeA).Sub(Time(sbTime))
					fmt.Println(fmt.Sprintf("%s %s %f %s",user.Name, dayStr, Binary(subM.Hours())-1, "小时"))
					overtime[day] = Binary(subM.Hours())
				}
			}



		} else {
			endStr := dayStr
			// 工作日加班,最晚时间-18:00
			if len(v) == 0{
				fmt.Println(fmt.Sprintf("%s %s %d %s",user.Name, dayStr, 0, "缺勤"))
				overtime[day] = AttendanceDeletion
				continue
			} else if len(v) == 1 && len(next) == 3 {
				// 今天打卡一次，明天打卡3次（加班到凌晨导致）
				fmt.Println(fmt.Sprintf("%s %s %s",user.Name, dayStr, "加班到凌晨"))
				attendance[day] = append(attendance[day], next[0])
				attendance[day+1] = []string{next[1], next[2]}
				endStr = fmt.Sprintf("%02d", day+1)
			} else if len(v) == 1 && len(next) != 3 {
				fmt.Println(fmt.Sprintf("%s %s %s",user.Name, dayStr, "考勤异常"))
				overtime[day] = AttendanceAbnormal
				continue
			}
			//begin := v[0]
			end := attendance[day][len(attendance[day])-1]
			time := fmt.Sprintf("%s-%s-%s %s", year, month, endStr, end)
			btime := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "18:00")
			etime := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "18:20")

			// 18:20后算加班
			if  Time(time).After(Time(etime)){
				subM := Time(time).Sub(Time(btime))
				fmt.Println(fmt.Sprintf("%s %s %f %s",user.Name, dayStr, Binary(subM.Hours()), "小时"))
				overtime[day] = Binary(subM.Hours())

			} else {
				fmt.Println(fmt.Sprintf("%s %s %d %s",user.Name, dayStr, 0, "小时"))
				overtime[day] = 0
			}

		}

	}

	return overtime
}

func checkHoliday(day int)  (bool, error){
	//dir, _ := os.Getwd()
	//path := dir+"/data/"+ month +".json"
	jsonFile, err := os.Open(ConfName)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err!= nil {
		log.Fatal(err)
		return false, err
	}
	var data []int
	json.Unmarshal(jsonData, &data)

	for _, v := range data{
		if v == day {
			return true, err
		}
	}
	return false, err
}

func Time(v string)  time.Time {
	local, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02 15:04", v, local)
	return t
}

// 小数后一位，满3变5，满8进1
func Binary(f float64) float64  {
	//str := fmt.Sprintf("%.1f", f)
	str := Truncate(f,2)
	if len(strings.Split(str, ".")) < 2 {
		fmt.Print(str)
		str += ".00"
	}
	xs0 := strings.Split(str, ".")[0]
	xs1 := strings.Split(str, ".")[1]
	if len(xs1) == 1 {
		xs1 += "0"
	}
	ahead, _ := strconv.Atoi(xs0)
	behind, _ := strconv.Atoi(xs1)

	if 33 <= behind && behind < 83 {
		behind = 5
	} else if behind >= 83 {
		ahead += 1
		behind = 0
	} else if behind < 33 {
		behind = 0
	}
	s := fmt.Sprintf("%d.%d", ahead, behind)
	re, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return re
}

func createExcel(list []model.User, folder string) (string,error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	index, err := f.NewSheet("加班统计")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	_ = f.SetColWidth("加班统计", "A", "AF", 10)


	rowNum := 1
	res := []interface{}{
		"姓名",
		"工号",
		" ",
	}
	for i :=1;i<=31;i++{
		res = append(res, i)
	}
	f.SetSheetRow("加班统计", fmt.Sprintf("A%d", rowNum),&res)
	for _, v := range list {
		rowNum ++
		res := []interface{}{
			v.Name,
			v.Number,
			"",
		}
		for j,_ := range v.Overtime {
			if j < 1 {
				continue
			}
			if v.Overtime[j] == AttendanceAbnormal {
				res = append(res, "异常")
			} else if  v.Overtime[j] == AttendanceDeletion {
				res = append(res, "缺勤")
			} else {
				res = append(res, v.Overtime[j])
			}
		}
		f.SetSheetRow("加班统计", fmt.Sprintf("A%d", rowNum),&res)
		rowNum ++
		f.SetSheetRow("加班统计", fmt.Sprintf("A%d", rowNum),&[]interface{}{})
	}



	f.SetActiveSheet(index)
	//dir, _ := os.Getwd()
	path := folder+fmt.Sprintf("/%d", time.Now().UnixMicro())+".xlsx"
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
		return path, err
	}

	return path, err
}

func Truncate(f float64, prec int) string {
	n := strconv.FormatFloat(f, 'f', -1, 64)
	if n == "" {
		return ""
	}
	if prec >= len(n) {
		return n
	}
	newn := strings.Split(n, ".")
	if len(newn) < 2 || prec >= len(newn[1]) {
		return n
	}
	return newn[0] + "." + newn[1][:prec]
}

func getHolidayShangBan(t, dayStr string) string {
	stime := fmt.Sprintf("%s-%s-%s %s", year, month, dayStr, "8:30")
	if  Time(t).After(Time(stime)){
		return t
	}
	return stime
}


