package main

import (
  "io"
  "os"
  "fmt"
  "log"
  "net/http"
  "html/template"
  "path/filepath"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

func init() {
  tpl = template.Must(template.ParseGlob("static/*.html"))

  db, dbErr = sql.Open("mysql", "anshuman:anshuman32@tcp(localhost:3306)/goDb")
  checkErr(dbErr)
  fmt.Println("Database connection successful...")
}

//-------- GLOBAL VARIABLES --------------
var adminGlobalEmail string
var adminGlobalPass string
var empFilePath string = "data/EmployeeMaster.json"
var db *sql.DB
var dbErr error
var tpl *template.Template
//----------------------------------------

//-------- TEMPLATES ---------------------
type UpdateData struct {
  Updater string
  Updatee Employee
}

type Dashboard struct {
  AllEmps []Employee
  AdminCreds Admin
  TotalEmps int
}

type Employee struct {
  Name string
  Email string
  Password string
  Position string
  Profile string
}

type Admin struct {
  Email, Password string
}
//----------------------------------------

// ----------- HELPER FUNCTIONS ----------
func checkErr(err error) {
  if err != nil {
    fmt.Println(err)
  }
}
func getEmployees() []Employee {
  var allEmps []Employee

  rows, empsErr := db.Query("select * from Employees")
  if empsErr != nil {
  }
  for rows.Next() {
    var emp Employee
    rows.Scan(&emp.Name, &emp.Email, &emp.Password, &emp.Position, &emp.Profile)
    allEmps = append(allEmps, emp)
  }
  return allEmps
}

func save(emp Employee) {
  //save to db
  _, insertErr := db.Query("insert into Employees (name, email, password, position, profile) values ('"+emp.Name+"','"+emp.Email+"','"+emp.Password+"','"+emp.Position+"','"+emp.Profile+"')")

  if insertErr != nil {
    fmt.Println(insertErr)
  }
  fmt.Println(emp, "Inserted!")
}
// ---------------------------------------


// ----------- ROUTE HANDLERS ----------
func admin(res http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    tpl.ExecuteTemplate(res, "admin.html", nil)
  }

  if req.Method == "POST" {
    adminGlobalEmail, adminGlobalPass = req.FormValue("email"), req.FormValue("password")

    adminCreds := Admin{Email: adminGlobalEmail, Password: adminGlobalPass}
    dashboard := Dashboard{AllEmps: getEmployees(), AdminCreds: adminCreds, TotalEmps: len(getEmployees())}

    tpl.ExecuteTemplate(res, "dashboard.html", dashboard)
  }
}

func employee(res http.ResponseWriter, req *http.Request) {
  if req.Method == "GET"  {

    tpl.ExecuteTemplate(res, "reg.html", nil)
  }
  if req.Method == "POST" {
    name := req.FormValue("name")
    email := req.FormValue("email")
    pass := req.FormValue("password")
    pos := req.FormValue("pos")

    req.ParseMultipartForm(5 * 1024 * 1024)
    file, handler, _ :=  req.FormFile("profile")

    defer file.Close()

    //save img 
    imgPath := filepath.Join("profiles/", handler.Filename)
    dst, err := os.Create(imgPath)
     checkErr(err)
    defer dst.Close()
    io.Copy(dst, file)

    e := Employee{Name: name, Email: email, Password: pass, Position: pos, Profile: imgPath}
    save(e)

    tpl.ExecuteTemplate(res, "reg.html", nil)
  }
}

func showEmployees(res http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    allEmps := getEmployees()

    tpl.ExecuteTemplate(res, "reg.html", allEmps)
  }
}

func update(res http.ResponseWriter, req *http.Request) {
  email := req.URL.Query().Get("email")

  if req.Method == "GET" {
    //search employee based on email
    var empToUpdate Employee
    allEmps := getEmployees()

    for _, emp := range allEmps {
      if emp.Email == email {
        empToUpdate = emp
        break
      }
    }

    updateInfo := UpdateData{Updater: adminGlobalEmail, Updatee: empToUpdate}
    tpl.ExecuteTemplate(res, "update.html", updateInfo)
  }
}

func saveEmp(res http.ResponseWriter, req *http.Request) {
  oldEmail := req.URL.Query().Get("email")
  if req.Method == "POST" {
    newName := req.FormValue("name")
    newPass := req.FormValue("pass")
    newEmail := req.FormValue("email")
    newPosition := req.FormValue("pos")

    upd, updErr := db.Prepare("update Employees set name=?, email=?, password=?, position=? where email=?")
    checkErr(updErr)
    upd.Exec(newName, newEmail, newPass, newPosition, oldEmail)

    adminCreds := Admin{Email: adminGlobalEmail, Password: adminGlobalPass}
    dashboard := Dashboard{AllEmps: getEmployees(), AdminCreds: adminCreds, TotalEmps: len(getEmployees())}

    tpl.ExecuteTemplate(res, "dashboard.html", dashboard)
  }
}

func remove(res http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    email := req.FormValue("email")
    _, delErr := db.Query("delete from Employees where email='"+email+"'")
    checkErr(delErr)
    fmt.Println(email , "deletion successful")

    adminCreds := Admin{Email: adminGlobalEmail, Password: adminGlobalPass}
    dashboard := Dashboard{AllEmps: getEmployees(), AdminCreds: adminCreds, TotalEmps: len(getEmployees())}

    tpl.ExecuteTemplate(res, "dashboard.html", dashboard)
  }
}

func home(res http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    tpl.ExecuteTemplate(res, "home.html", nil)
  }
}
// ---------------------------------------

func main() {
  log.Println("Server is running....")

  defer db.Close()

  http.HandleFunc("/", home)
  http.HandleFunc("/admin", admin)
  http.HandleFunc("/employee/", employee)
  http.HandleFunc("/showEmployees", showEmployees)
  http.HandleFunc("/admin/remove/", remove)
  http.HandleFunc("/admin/update/", update)
  http.HandleFunc("/admin/save", saveEmp)

  // file server for images
  fs := http.StripPrefix("/resources/profiles", http.FileServer(http.Dir("./profiles")))
  http.Handle("/resources/", fs)

  http.ListenAndServe(":8000", nil)
}
