package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	uuid "github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	UserName  string
	Password  []byte
	FirstName string
	LastName  string
	Role      string
}

type Product struct {
	CusID   int
	CusName string
	Email   string
	Gender  string
	City    string
	Zipcode string
	Country string
}

var (
	templates = template.Must(template.ParseGlob("template/*"))
)

var db *sql.DB
var dbSession = map[string]string{}
var dbUser = map[string]user{}
var err error

/*
func init() {
	bs, err := bcrypt.GenerateFromPassword([]byte("charles"), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	dbUser["iykecharles316@yahoo.com"] = user{"iykecharles316@yahoo.com", bs, "charles", "ezema", "backend"}
}
*/

func main() {
	//connection
	var err error
	// Never use _, := db.Open(), resources need to be released with db.Close
	db, err = sql.Open("mysql", "root:charles@tcp(localhost:3306)/charlesdb")
	// psw := os.Getenv("mysql_password")
	// db, err = sql.Open("mysql", "root:"+psw+"@tcp(localhost:3306)/charlesdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", landingpage)
	http.HandleFunc("/create", create)
	http.HandleFunc("/droptable", droptable)
	http.HandleFunc("/vault", vault)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/insert/process", mainpage)
	http.HandleFunc("/update", update1)
	http.HandleFunc("/update/process", update2)
	http.HandleFunc("/delete", deleterow)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalln(err)
	}

}

func landingpage(w http.ResponseWriter, r *http.Request) {
	s := getUser(r)
	templates.ExecuteTemplate(w, "landing.html", s)
}

func create(w http.ResponseWriter, r *http.Request) {
	stmt, err := db.Prepare(`create table IF NOT EXISTS person (
		id SERIAL PRIMARY KEY,
		names VARCHAR(50) NOT NULL,
		email VARCHAR(150),
		gender VARCHAR(7) NOT NULL,
		city VARCHAR(7) NOT NULL,
		zipcode VARCHAR(7) NOT NULL,
		country_of_birth VARCHAR(50),
		constraint unique_email unique (email)

	);`)
	check(err)
	defer stmt.Close()

	res, err := stmt.Exec()
	check(err)

	n, err := res.RowsAffected()
	check(err)

	fmt.Fprintln(w, "CREATED A TABLE", n)
}

func droptable(w http.ResponseWriter, r *http.Request) {
	st := `DROP TABLE IF EXISTS person;`
	drpstmt, err := db.Prepare(st)
	check(err)

	defer drpstmt.Close()

	res, err := drpstmt.Exec()
	check(err)

	n, err := res.RowsAffected()
	check(err)

	fmt.Fprintln(w, "TABLE SUCCESFULLY DROPPED", n)

}

func deleterow(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.FormValue("id")
	delstmt, err := db.Prepare(`DELETE FROM person WHERE id = ?;`)
	if err != nil {
		fmt.Println(err)
	}

	defer delstmt.Close()

	res, err := delstmt.Exec(id)
	if err != nil {
		fmt.Println(err)
	}

	n, err := res.RowsAffected()

	fmt.Fprintln(w, "data deleted successfully", n)
	fmt.Println("deleted successfully")
	templates.ExecuteTemplate(w, "vault.html", "DELETED")

}

func insert(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "insert.html", nil)
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	if !alreadyloggedin(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s := getUser(r)
	if s.Role != "backend" {
		fmt.Println("User isnt a backend developer. Access denied")
		http.Error(w, "Access denied. You must be a backend developer to access this infromation", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		fmt.Println("get method")
		return
	}

	// INSERT PROCESS
	if r.Method == http.MethodPost {
		// Formvalues
		r.ParseForm()
		z := Product{}
		z.CusName = r.FormValue("customername")
		z.Email = r.FormValue("email")
		z.Gender = r.FormValue("gender")
		z.City = r.FormValue("city")
		z.Zipcode = r.FormValue("zipcode")
		z.Country = r.FormValue("country")

		if z.CusName == "" || z.Email == "" || z.Gender == "" || z.City == "" || z.Zipcode == "" || z.Country == "" {
			fmt.Println("Insert fields cant be blank")
			http.Error(w, "Insert Fields cant be blank. Return and fill all the spaces", http.StatusForbidden)
			return
		}

		// inserting values
		insertstmt := `INSERT IGNORE into person (names, email, gender, city, zipcode, country_of_birth) VALUES (?, ?, ?, ?, ?, ?)`
		_, err := db.Exec(insertstmt, z.CusName, z.Email, z.Gender, z.City, z.Zipcode, z.Country)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	templates.ExecuteTemplate(w, "index.html", "Product Successfully Inserted")
}

func update1(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	stmt := `SELECT * FROM charlesdb.person WHERE id = ?;`
	row, err := db.Query(stmt, id)
	if err != nil {
		http.Error(w, "couldnt update data", http.StatusNoContent)
		fmt.Println("couldnt update data")
		return
	}

	z := Product{}
	//var z Product

	for row.Next() {
		err := row.Scan(&z.CusID, &z.CusName, &z.Email, &z.Gender, &z.City, &z.Zipcode, &z.Country)
		check(err)
	}
	templates.ExecuteTemplate(w, "update.html", z)
}

func update2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("etttt")
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	z := Product{}
	id := r.FormValue("id")
	z.CusName = r.FormValue("customername")
	z.Email = r.FormValue("email")
	z.Gender = r.FormValue("gender")
	z.City = r.FormValue("city")
	z.Zipcode = r.FormValue("zipcode")
	z.Country = r.FormValue("country")

	// "UPDATE `testdb`.`products` SET `name` = ?, `price` = ?, `description` = ? WHERE (`idproducts` = ?);"
	// "UPDATE Employee SET name=?, city=? WHERE id=?"
	upstmt := "UPDATE person SET names = ?, email = ?, gender = ?, city = ?, zipcode = ?, country_of_birth = ? WHERE id = ?;"
	stmt, err := db.Prepare(upstmt)
	if err != nil {
		fmt.Println("error here", err)
	}

	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(z.CusName, z.Email, z.Gender, z.City, z.Zipcode, z.Country, id)
	rowsAff, _ := res.RowsAffected()
	if err != nil || rowsAff != 1 {
		fmt.Println(err)
		templates.ExecuteTemplate(w, "update2.html", "There was a problem updating the product")
		return
	}

	/*
		if err != nil {
			fmt.Println("error here2 :", err)
		}
	*/

	fmt.Println("Data Successfully Updated")
	fmt.Fprintln(w, "Data Successfully Updated")
	templates.ExecuteTemplate(w, "update2.html", "Product was Successfully Updated")
}

func vault(w http.ResponseWriter, r *http.Request) {

	smt := `SELECT * FROM person`
	rows, err := db.Query(smt)
	check(err)

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var z Product
		err = rows.Scan(&z.CusID, &z.CusName, &z.Email, &z.Gender, &z.City, &z.Zipcode, &z.Country)
		check(err)

		products = append(products, z)
	}

	templates.ExecuteTemplate(w, "vault.html", products)
}

func signup(w http.ResponseWriter, r *http.Request) {
	if alreadyloggedin(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// formvalues
		un := r.FormValue("username")
		p := r.FormValue("password")
		f := r.FormValue("firstname")
		l := r.FormValue("lastname")
		ro := r.FormValue("role")

		if un == "" || p == "" || f == "" || l == "" || ro == "" {
			fmt.Println("Signup fields cant be blank")
			http.Error(w, "Signup Fields cant be blank. Return and fill all the spaces", http.StatusForbidden)
			return
		}

		// session
		id, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}
		c := &http.Cookie{
			Name:  "charlescookie",
			Value: id.String(),
			Path:  "/",
		}
		http.SetCookie(w, c)

		// username exist?
		if _, ok := dbUser[un]; ok {
			http.Error(w, "username exist already", http.StatusForbidden)
			return
		}

		// interaction
		// encrypt password using bcrypt
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		dbSession[c.Value] = un
		u := user{un, bs, f, l, ro}
		dbUser[un] = u

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return

	}

	templates.ExecuteTemplate(w, "signup.html", nil)
}

func login(w http.ResponseWriter, r *http.Request) {

	if !alreadyloggedin(r) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		// formvalues
		un := r.FormValue("username")
		p := r.FormValue("password")

		// check if user exist
		u, ok := dbUser[un]
		if !ok {
			http.Error(w, "username deosnt match", http.StatusForbidden)
			return
		}

		// password exist?
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "passwords do not match", http.StatusForbidden)
			return
		}

		// create a session
		id, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}
		c := &http.Cookie{
			Name:  "charlescookie",
			Value: id.String(),
			Path:  "/",
		}
		http.SetCookie(w, c)

		// the interaction
		dbSession[c.Value] = un
		dbUser[un] = u

		http.Redirect(w, r, "/insert", http.StatusSeeOther)
		return

	}
	templates.ExecuteTemplate(w, "login.html", nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if alreadyloggedin(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// delete session
	c, _ := r.Cookie("charlescookie")
	delete(dbSession, c.Value)

	// deactivate cookie so value is emply string and MaxAge = -1
	c = &http.Cookie{
		Name:   "charlescookie",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return

}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
