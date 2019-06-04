package main

import (
	h "./handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main () {
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/main/:username/*password", h.SalaryLastWeek)
	router.GET("/sum-of-years/:username/*password", h.SumOfYearSalary)
	router.GET("/send-money/:data", h.SendMoneyHandler) //username="usr"&password="pswd"&id="2013232"&amount="5450"
	router.GET("/current-money/:username/*password", h.CurrentMoneyHandler)
	router.Run(":8080")
}