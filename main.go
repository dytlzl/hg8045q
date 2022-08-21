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
	defer Logout(ctx)
	err = GetWanList(ctx)
	if err != nil {
		panic(err)
	}
	err = GetLanUserDevInfo(ctx)
	if err != nil {
		panic(err)
	}
	err = dhcpStaticIPConfigs(ctx)
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

type dhcpConfig struct {
	Domain     string
	IsEnabled  bool
	IP         string
	MACAddress string
}

func dhcpStaticIPConfigs(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/html/bbsp/dhcpstatic/dhcpstatic.asp", nil)
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
	key := "new stDhcp("
	isFound := false
	data := make([]dhcpConfig, 0)
	for i := range result {
		if strings.HasPrefix(result[i:], key) {
			start = i + len(key)
			isFound = true
		}
		if isFound && strings.HasPrefix(result[i:], "),") {
			row := result[start:i]
			cells := strings.Split(row, ",")
			data = append(data, dhcpConfig{
				Domain:     strings.Trim(cells[0], "\""),
				IsEnabled:  cells[1] == "1",
				IP:         strings.Trim(cells[2], "\""),
				MACAddress: strings.Trim(cells[3], "\""),
			})
			isFound = false
		}
	}
	PrintTable([]string{"STATIC IP", "MAC ADDRESS"}, data, func(d dhcpConfig) []string {
		return []string{d.IP, d.MACAddress}
	}, "")
	return nil
}

func GetLanUserDevInfo(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/html/bbsp/common/GetLanUserDevInfo.asp", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := string(bytes)
	start := 0
	key := "new USERDevice("
	isFound := false
	data := make([]UserDevice, 0)
	for i := range result {
		if strings.HasPrefix(result[i:], key) {
			start = i + len(key)
			isFound = true
		}
		if isFound && strings.HasPrefix(result[i:], "),") {
			row := result[start:i]
			cells := strings.Split(row, ",")
			data = append(data, UserDevice{
				IP:         strings.Trim(cells[1], "\""),
				MACAddress: strings.Trim(cells[2], "\""),
				PortID:     strings.Trim(cells[3], "\""),
				Status:     strings.Trim(cells[6], "\""),
				Hostname:   strings.Trim(cells[9], "\""),
			})
			isFound = false
		}
	}
	PrintTable([]string{"IP", "MAC ADDRESS", "PORT ID", "STATUS", "HOSTNAME"}, data, func(d UserDevice) []string {
		return []string{d.IP, d.MACAddress, d.PortID, d.Status, d.Hostname}
	}, "")
	return nil
}

func GetWanList(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/html/bbsp/common/wan_list.asp", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := string(bytes)
	start := 0
	key := "new WanIP("
	isFound := false
	for i := range result {
		if strings.HasPrefix(result[i:], key) {
			start = i + len(key)
			isFound = true
		}
		if isFound && strings.HasPrefix(result[i:], "),") {
			row := result[start:i]
			cells := strings.Split(row, ",")
			isFound = false
			fmt.Println("GLOBAL IP:", strings.Trim(cells[12], "\""))
			return nil
		}
	}
	return nil
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
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Cookie":       {fmt.Sprintf("Cookie=UserName:%s:PassWord:%s:Language:japanese:id=-1", username, base64.StdEncoding.EncodeToString([]byte(password)))},
	}
	resp, err := client.Do(request)
	if len(resp.Header["Set-Cookie"]) == 0 {
		return errors.New("login failed")
	}
	if err != nil {
		return err
	}

	request, err = http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/frame.asp", nil)
	if err != nil {
		return err
	}
	resp, err = client.Do(request)
	if err != nil {
		return err
	}
	return nil
}

func Logout(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://192.168.1.1/logout.cgi", nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
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
