package handlers

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseURL = "https://www.alivemax.com"
)

type App struct {
	Client *http.Client
}

type YearInfo struct {
	Year string
	Date string
	Amount string
}

type YearInfos struct {
	arr []YearInfo
	Sum18 float64
	Sum17 float64
	Sum16 float64
}

type CurrentData struct {
	money string
	fullName string
}

type Name struct {
	name string
}

type Salary struct {
	Date string
	Amount string
}

func (app *App) login(username, password string) {
	client := app.Client

	loginURL := baseURL + "/login"

	data := url.Values{
		"LOGIN": {"1"},
		"USER":  {username},
		"PASS":  {password},
	}

	response, err := client.PostForm(loginURL, data)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
}
func (app *App) sendMoney(id, amount, password string) {
	client := app.Client
	page := url.Values{
		"STP": {"3"}, //= 3  ??? int or string
		"AMOUNT": {amount}, //= 0.50
		"PASSWORD": {password}, //=
		"GET_USER": {""},
		"GET_ID": {id},
		"FOUND_ID": {id},
		"ID": {id}, //= id
	}
	response, err := client.PostForm(baseURL + "/transferCredit", page)

	if err != nil {
		log.Fatal("error", err)
	}
	defer response.Body.Close()
	fmt.Println("Successfully: ", response.Status)

}

func SendMoneyHandler(c *gin.Context) {
	data := c.Param("data")
	fromUrl, err := url.ParseQuery(data)
	if err != nil {
		return
	}
	username := fromUrl["username"][0]
	password := fromUrl["password"][0]
	id := fromUrl["id"][0]
	amount := fromUrl["amount"][0]
	jar, _ := cookiejar.New(nil)

	app := App {
		Client: &http.Client{Jar: jar},
	}

	app.login(username, password)
	app.sendMoney(id, amount, password)
}

func (app *App) getNextPage(i int) []Salary {
	client := app.Client
	page := url.Values{
		"SCROLL::PAGE": {strconv.Itoa(i)},
	}
	response, err := client.PostForm(baseURL+"/earnings", page)

	if err != nil {
		log.Fatal("error", err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}


	var salary []Salary

	document.Find(".mout").Each(func(i int, s *goquery.Selection) {
		date := s.Find("td").Slice(1, 2).Text()
		amount := s.Find("td").Slice(2, 3).Text()

		result := Salary{
			Date: date,
			Amount: amount,
		}
		salary = append(salary, result)
	})
	return salary
}

func SalaryLastWeek(c *gin.Context) {
	username := c.Param("username")
	password := c.Param("password")[1:]

	jar, _ := cookiejar.New(nil)

	app := App{
		Client: &http.Client{Jar: jar},
	}

	app.login(username, password)
	a := app.getNextPage(0)

	c.JSON(200, gin.H{
		"lastWeekSalary": a[0].Amount,
		"lastWeekDate": a[0].Date,
	})
}


func SumOfYearSalary (c *gin.Context) {
	username := c.Param("username")
	password := c.Param("password")[1:]

	jar, _ := cookiejar.New(nil)

	app := App{
		Client: &http.Client{Jar: jar},
	}
	//var year YearInfos
	app.login(username, password)

	var year YearInfos

	for i := 2; i <= 9; i++ {
		a := app.getNextPage(i)
		for _, b := range a {
			if strings.Contains(b.Date, "2018") {
				y := YearInfo{"2018", b.Date, b.Amount}
				year.arr = append(year.arr, y)
			} else if strings.Contains(b.Date, "2017") {
				y := YearInfo{"2017", b.Date, b.Amount}
				year.arr = append(year.arr, y)
			} else if strings.Contains(b.Date, "2016") {
				y := YearInfo{"2016", b.Date, b.Amount}
				year.arr = append(year.arr, y)
			}
		}
	}

	for _, b := range year.arr {
		if b.Year == "2018" {
			conv, _ := strconv.ParseFloat(b.Amount[1:], 64)
			year.Sum18 += conv
		} else if b.Year == "2017" {
			conv, _ := strconv.ParseFloat(b.Amount[1:], 64)
			year.Sum17 += conv
		} else if b.Year == "2016" {
			conv, _ := strconv.ParseFloat(b.Amount[1:], 64)
			year.Sum16 += conv
		}
	}

	c.JSON(200, gin.H{
		"2018": year.Sum18,
		"2017": year.Sum17,
		"2016": year.Sum16,
	})

}

func (app *App) GetCurrentMoney() CurrentData {
	client := app.Client
	salaryURL := baseURL + "/home"

	response, err := client.Get(salaryURL)

	if err != nil {
		log.Fatal("[GetCurrentMoney] error", err)
	}
	defer response.Body.Close()

	page, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("[GetCurrentMoney] Error loading HTTP response body", err)
	}

	var currentData CurrentData
	moneyLine := page.Find("font").Text()
	if len(moneyLine) == 74 {
		currentData.money = moneyLine[9:18]
	} else if len(moneyLine) == 73 {
		currentData.money = moneyLine[9:17]
	} else if len(moneyLine) == 72 {
		currentData.money = moneyLine[9:16]
	} else if len(moneyLine) == 71 {
		currentData.money = moneyLine[9:15]
	} else if len(moneyLine) == 70 {
		currentData.money = moneyLine[9:12]
	} else {
		currentData.money = "нет денег"
	}

	//var name Name
	currentData.fullName = page.Find("td.nameBanner").Text()
	return currentData
}

func CurrentMoneyHandler(c *gin.Context) {
	username := c.Param("username")
	password := c.Param("password")[1:]

	jar, _ := cookiejar.New(nil)

	app := App {
		Client: &http.Client{Jar: jar},
	}

	app.login(username, password)
	result := app.GetCurrentMoney()

	c.JSON(200, gin.H{
		"money": result.money,
		"fullName": result.fullName,
	})

}