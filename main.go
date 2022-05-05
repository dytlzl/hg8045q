package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

var client *http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
	randCount, err := GetRandCount(ctx)
	if err != nil {
		panic(err)
	}
	err = Login(ctx, randCount)
	if err != nil {
		panic(err)
	}
	err = GetLanUserDevInfo(ctx)
	if err != nil {
		panic(err)
	}
}

type UserDevice struct {
	IP         string
	MACAddress string
	Status     string
	Hostname   string
	PortID     string
}

func GetLanUserDevInfo(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/html/bbsp/common/GetLanUserDevInfo.asp", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := string(bytes)
	start := 0
	key := "new USERDevice("
	data := make([]UserDevice, 0)
	for i := range result {
		if strings.HasPrefix(result[i:], key) {
			start = i + len(key)
		}
		if strings.HasPrefix(result[i:], "),") {
			row := result[start:i]
			cells := strings.Split(row, ",")
			data = append(data, UserDevice{
				IP:         LRStrip(cells[1]),
				MACAddress: LRStrip(cells[2]),
				PortID:     LRStrip(cells[3]),
				Status:     LRStrip(cells[6]),
				Hostname:   LRStrip(cells[9]),
			})
		}
	}
	PrintTable([]string{"IP", "MAC ADDRESS", "PORT ID",  "STATUS", "HOSTNAME"}, data, func(d UserDevice) []string {
		return []string{d.IP, d.MACAddress, d.PortID, d.Status, d.Hostname}
	}, "")
	return nil
}

func LRStrip(s string) string {
	return s[1 : len(s)-1]
}

func PrintTable[T any](columnNames []string, data []T, fn func(T) []string, prefix string) {
	table := make([][]string, len(data))

	columnWidthList := make([]int, len(columnNames))
	for x := range columnWidthList {
		columnWidthList[x] = len(columnNames[x])
	}
	for index, element := range data {
		table[index] = fn(element)
		for x := range columnWidthList {
			if columnWidthList[x] < len(table[index][x]) {
				columnWidthList[x] = len(table[index][x])
			}
		}
	}

	for x, cell := range columnNames {
		fmt.Printf("%s%s%s", prefix, cell, strings.Repeat(" ", columnWidthList[x]+3-len(cell)))
	}
	fmt.Println()
	for _, row := range table {
		for x, cell := range row {
			fmt.Printf("%s%s%s", prefix, cell, strings.Repeat(" ", columnWidthList[x]+3-len(cell)))
		}
		fmt.Println()
	}
}

func Login(ctx context.Context, randCount string) error {
	form := url.Values{
		"x.X_HW_Token": {randCount},
	}
	request, err := http.NewRequestWithContext(ctx, "POST", "http://192.168.1.1/login.cgi", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	username, ok := os.LookupEnv("HG8045Q_USERNAME")
	if !ok {
		return errors.New("environment variable 'HG8045Q_USERNAME' is not set")
	}
	password, ok := os.LookupEnv("HG8045Q_PASSWORD")
	if !ok {
		return errors.New("environment variable 'HG8045Q_PASSWORD' is not set")
	}
	request.Header = http.Header{
		"Content-Type":              {"application/x-www-form-urlencoded"},
		"Cookie":                    {fmt.Sprintf("Cookie=UserName:%s:PassWord:%s:Language:japanese:id=-1", username, base64.StdEncoding.EncodeToString([]byte(password)))},
	}
	resp, err := client.Do(request)
	if len(resp.Header["Set-Cookie"]) == 0 {
		return errors.New("login failed")
	}
	if err != nil {
		return err
	}
	return nil
}

func GetRandCount(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/asp/GetRandCount.asp", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// strip leading BOM 239 187 191
	return string(bytes[3:]), nil
}
