package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"

	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() {
	dbSave(p)
	return
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func dbSave(p *Page) {
	db, err := sql.Open("mysql", "root:admin@/gowiki")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	stmtOut, err := db.Prepare("SELECT 1 FROM page WHERE title = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var ret int
	fmt.Println("SELECT")
	err = stmtOut.QueryRow(p.Title).Scan(&ret) // WHERE number = 13
	if err != nil {
		// TODO no record has err value, but it should not occuer panic
		// fmt.Println(err.Error())
		//panic(err.Error()) // proper error handling instead of panic in your app
	}

	if ret >= 1 {
		stmtDel, err := db.Prepare("DELETE FROM page WHERE title = ?")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtDel.Close()

		fmt.Println("DELETE")

		err = stmtDel.QueryRow(p.Title).Scan(&ret)
		if err != nil {
      // TODO sql: no rows in result set goroutine, even there is row.
      // need to check the cause
			//panic(err.Error()) // proper error handling instead of panic in your app
		}
	}

	fmt.Println("INSERT")

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO page VALUES( ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(p.Title, p.Body)
	if err != nil {
		panic(err.Error())
	}
}

func dbSelect(title string) (*Page, error) {
	db, err := sql.Open("mysql", "root:admin@/gowiki")
	fmt.Printf(" type %T", db)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	stmtOut, err := db.Prepare("SELECT body FROM page WHERE title = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var savedBody []byte
	fmt.Println("SELECT")
	err = stmtOut.QueryRow(title).Scan(&savedBody) // WHERE number = 13
	fmt.Println(savedBody)
	if err != nil {
		//f
		//panic(err.Error()) // proper error handling instead of panic in your app
	}
	// fmt.Printf("The ret  is: %d", ret)
	if savedBody != nil {
		return &Page{Title: title, Body: savedBody}, nil
	} else {
		fmt.Println("No res savedBody")
		return nil, err
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8888", nil)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	//p, err := loadPage(title)
	p, err := dbSelect(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	//p, err := loadPage(title)
	p, err := dbSelect(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {

	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
