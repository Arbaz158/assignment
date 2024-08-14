package models

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	employee "github.com/Employee"
	"github.com/config"
	"github.com/xuri/excelize/v2"
)

var expectedHeaders = []string{
	"first_name",
	"last_name",
	"company_name",
	"address",
	"city",
	"county",
	"postal",
	"phone",
	"email",
	"web",
}

func ProcessFile(filename string) []employee.Employee {
	defer os.Remove(filename)

	f, err := excelize.OpenFile(filename)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil
	}

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		log.Println("Error reading rows:", err)
		return nil
	}

	var wg sync.WaitGroup
	dataChannel := make(chan employee.Employee)

	if len(rows) > 0 && !ValidateRow(rows[0]) {
		log.Println("Invalid header row")
		return nil
	}

	for _, row := range rows[1:] {
		wg.Add(1)
		go func(row []string) {
			defer wg.Done()
			if len(row) == len(expectedHeaders) {
				data := employee.Employee{
					FirstName:   row[0],
					LastName:    row[1],
					CompanyName: row[2],
					Address:     row[3],
					City:        row[4],
					Country:     row[5],
					Postal:      row[6],
					Phone:       row[7],
					Email:       row[8],
					Web:         row[9],
				}
				dataChannel <- data
			}
		}(row)
	}

	go func() {
		wg.Wait()
		close(dataChannel)
	}()

	var employees []employee.Employee
	for data := range dataChannel {
		employees = append(employees, data)
	}

	return employees
}

func ValidateRow(row []string) bool {
	if len(row) != len(expectedHeaders) {
		return false
	}

	for i, header := range expectedHeaders {
		if row[i] != header {
			return false
		}
	}
	return true
}

func CreateEmployeeTable() error {
	db := config.Dbconnect()
	_, err := db.Exec(`CREATE TABLE employees (
                        id INT AUTO_INCREMENT PRIMARY KEY,
                        first_name VARCHAR(255) NOT NULL,
                        last_name VARCHAR(255) NOT NULL,
                        company_name VARCHAR(255),
                        address VARCHAR(255),
                        city VARCHAR(255),
                        country VARCHAR(255),
                        postal VARCHAR(20),
                        phone VARCHAR(20),
                        email VARCHAR(255),
                        web VARCHAR(255)
                    );
                `)
	if err != nil {
		log.Println("error in creating table :", err)
		return err
	}
	return nil
}

func InsertEmployee(emp employee.Employee) error {
	db := config.Dbconnect()
	stmt, err := db.Prepare("INSERT INTO employees (first_name,last_name,company_name,address,city,country,postal,phone,email,web) VALUES (?,?,?,?,?,?,?,?,?,?) ")
	if err != nil {
		log.Println("error in preparing statement :", err)
		return err
	}
	_, err = stmt.Exec(emp.FirstName, emp.LastName, emp.CompanyName, emp.Address, emp.City, emp.Country, emp.Postal, emp.Phone, emp.Email, emp.Web)
	if err != nil {
		log.Println("error in executing query :", err)
		return err
	}
	return nil
}

func CacheEmployees(employees []employee.Employee) error {
	redisClient := config.InitRedis()
	defer redisClient.Close()

	employeesJSON, err := json.Marshal(employees)
	if err != nil {
		return err
	}

	err = redisClient.Set("employees", employeesJSON, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetCachedData(key string) ([]employee.Employee, error) {
	var empCachedData []employee.Employee
	rdb := config.InitRedis()
	cachedData, err := rdb.Get(key).Result()
	if err != nil {
		log.Println("error in getting cached data :", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(cachedData), &empCachedData)
	if err != nil {
		log.Println("error in unmarshalling data :", err)
		return nil, err
	}
	return empCachedData, nil
}

func GetEmployeeData() ([]employee.Employee, error) {
	var empData []employee.Employee
	db := config.Dbconnect()
	rows, err := db.Query("SELECT * FROM employees")
	if err != nil {
		log.Println("error in query :", err)
		return nil, err
	}
	for rows.Next() {
		var emp employee.Employee
		err := rows.Scan(&emp.FirstName, &emp.LastName, &emp.CompanyName, &emp.Address, &emp.City, &emp.Country, &emp.Postal, &emp.Phone, &emp.Email, &emp.Web)
		if err != nil {
			log.Println("error in scanning employee data :", err)
			return nil, err
		}
		empData = append(empData, emp)
	}
	return empData, nil
}
