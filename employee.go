// Copyright 2016 Tahir Rauf. All rights reserved.
// This code is under Open Source license and anyone can use it 
// Author: Tahir Rauf <tahir.abdul.rauf@gmail.com>

package employee

import (
	"html/template"
	"net/http"
	"time"
	"appengine"
	"appengine/datastore"
        "appengine/memcache"
)

// This is the struct I will use for Rapporr Employees data 
type Employee struct {
        EmployeeNumber string
        FirstName  string
        LastName string
        DateInserted    time.Time
}

func init() {
	http.HandleFunc("/", listEmployees)
	http.HandleFunc("/add_employee", addEmployee)
        http.HandleFunc("/view_employee", viewEmployee) 
        http.HandleFunc("/save_employee", saveEmployee)
}

// key for all entries
func employeeKey(context appengine.Context) *datastore.Key {
	return datastore.NewKey(context, "Employee", "employee_first_store1", 0, nil)
}

// Instead of hardcoding the HTML, I found a way to create the HTML templates for better separation of concerns
var mainPage = template.Must(template.ParseFiles("templates/home.html", 
                    "templates/add_employee.html",
                    "templates/view_employee.html",
                    "templates/custom_error.html"))

// This handler is used to list all the employees
func listEmployees(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, "GET requests only", http.StatusMethodNotAllowed)
		return
	}
	if request.URL.Path != "/" {
		http.NotFound(responseWriter, request)
		return
	}

	context := appengine.NewContext(request)
	query := datastore.NewQuery("Employee").Ancestor(employeeKey(context)).Order("-DateInserted")
	var employee []*Employee
	if _, err := query.GetAll(context, &employee); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mainPage.ExecuteTemplate(responseWriter,"home.html", employee); err != nil {
		context.Errorf("%v", err)
	}
}

// This handler is used to redirect user to the template 
// so that input can be taken for employee data for addition
func addEmployee(responseWriter http.ResponseWriter, request *http.Request) {
    
	context := appengine.NewContext(request)
	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mainPage.ExecuteTemplate(responseWriter,"add_employee.html", nil); err != nil {
		context.Errorf("%v", err)
	}
}

// This handler is used to save the html form data to the datastore
func saveEmployee(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(responseWriter, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	context := appengine.NewContext(request)
	employee := &Employee{
		EmployeeNumber: request.FormValue("empnumb"),
                FirstName: request.FormValue("fname"),
                LastName: request.FormValue("lname"),
		DateInserted:    time.Now(),
	}
        
	key := datastore.NewIncompleteKey(context, "Employee", employeeKey(context))
	if _, err := datastore.Put(context, key, employee); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(responseWriter, request, "/", http.StatusSeeOther)
}


// This handler is used to view details of a particular employee
func viewEmployee(responseWriter http.ResponseWriter, request *http.Request) {
        context := appengine.NewContext(request)
        EmpNumb := request.URL.Query().Get("empnumb")

        // To check if user did not enter any query string
	if EmpNumb != "" {
        
            var emp Employee
            
            // I will first check the Employee data in Memcache, if its a miss
            // then I will query the data store and then store it in the memcache
            // otherwise use the memcache data
            _, erro:= memcache.Gob.Get(context, EmpNumb, &emp)

            if erro == memcache.ErrCacheMiss{

                query := datastore.NewQuery("Employee").Ancestor(employeeKey(context)).Filter("EmployeeNumber=",EmpNumb).Order("-DateInserted")
                var employee []*Employee

                if _, err := query.GetAll(context, &employee); err != nil {
                        http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
                        return
                    } 

                // In case of getting a successful employee against its Employee Number, render it!
                if (len(employee) != 0){
               
                    // In case of a Miss, I will add the object to the memcache to increase
                    // retrieval speed.
                    item:= &memcache.Item{
                    Key: EmpNumb,
                    Object: employee[0]}

                    memcache.Gob.Set(context, item)
                 
                    responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
                        if err := mainPage.ExecuteTemplate(responseWriter,"view_employee.html", employee[0]); err != nil {
                                context.Errorf("%v", err)
                    }
                // In case of invalid Employee Number, Show error page
                }else{
                    responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
                                    if err := mainPage.ExecuteTemplate(responseWriter,"custom_error.html", "Invalid Employee Number added in query"); err != nil {
                                            context.Errorf("%v", err)
                                    }
                    }
            // In case of empty query string, Show error page
            }else{
                    responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
                        if err := mainPage.ExecuteTemplate(responseWriter,"view_employee.html", emp); err != nil {
                                context.Errorf("%v", err)
                    }                    

                }
        }else{
                responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
                                    if err := mainPage.ExecuteTemplate(responseWriter,"custom_error.html", "No Employee Number added in query"); err != nil {
                                            context.Errorf("%v", err)
                                    }
        }
	

}