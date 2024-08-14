package controller

import (
	"log"
	"net/http"
	"sync"

	employee "github.com/Employee"
	"github.com/gin-gonic/gin"
	"github.com/models"
)

func UploadExcelAndFormatData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file"})
		return
	}

	dataChannel := make(chan []employee.Employee)

	go func() {
		employees := models.ProcessFile(file.Filename)
		dataChannel <- employees
	}()

	employees := <-dataChannel
	log.Println("Employee data:", employees)

	err = models.CreateEmployeeTable()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create table"})
		return
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(employees))
	cache_errors := make(chan error, len(employees))

	for _, emp := range employees {
		wg.Add(1)
		go func(emp employee.Employee) {
			defer wg.Done()
			if err := models.InsertEmployee(emp); err != nil {
				errors <- err
			}
		}(emp)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := models.CacheEmployees(employees); err != nil {
			//log.Println("Error in caching data into Redis:", err)
			cache_errors <- err
		}
	}()

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		for err := range errors {
			log.Println("Error while inserting data:", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Some data could not be inserted"})
		return
	}

	if len(cache_errors) > 0 {
		for err := range errors {
			log.Println("Error while caching data:", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Some data could not be cached"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Uploaded file, parsed it successfully.", "employees": employees})
}

func GetMysqlAndRedisData(c *gin.Context) {
	employees, err := models.GetCachedData("employees")
	if err != nil {
		log.Println("error in fetching cached data :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error in fetching cached data"})
	}

	if employees == nil {
		employees, err = models.GetEmployeeData()
		if err != nil {
			log.Println("error in fetching employee data :", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error in fetching employee data"})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "employee data list", "employees": employees})
}
