package model

import "strconv"

type User struct {
	Name       string     `json:"name"`
	Number     string     `json:"number"`
	Attendance [][]string `json:"attendance"`
	Overtime   []float64     `json:"overtime"`
}

type UserList []User

func (x UserList) Len() int {
	return len(x)
}
func (x UserList) Less(i, j int) bool {
	l, _ := strconv.Atoi(x[i].Number)
	b, _ := strconv.Atoi( x[j].Number)
	return l <b
}
func (x UserList) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}
