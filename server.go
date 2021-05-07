package main

import (
	"context"
	"encoding/json"
	_ "github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
	"github.com/rs/cors"
	"log"
	"net/http"
	"time"
)

type Server struct {
	handler         http.Handler
	db              *kivik.DB
	listenAddr      string
	couchdbUrl      string
	couchdbName     string
	dbAccessString  string
	googleSheetCode string
	frontPage       []PageSection
	lastPageRefresh time.Time
}

func NewServer(la, cdb, dbn, gsc string) *Server {
	a := Server{}
	a.listenAddr = la
	a.couchdbUrl = cdb
	a.couchdbName = dbn
	a.googleSheetCode = gsc
	a.dbAccessString = a.couchdbUrl + "/" + a.couchdbName

	sm := http.NewServeMux()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	// Required handlers
	sm.HandleFunc("/login", a.LoginHandler)
	sm.HandleFunc("/createauth", a.CreateAuthHandler)
	sm.HandleFunc("/genpendingauth", a.GenPendingAuthHandler)

	// Helper handlers
	sm.HandleFunc("/hash", a.HashHandler)
	sm.HandleFunc("/timestamp", a.TimestampHandler)
	sm.HandleFunc("/id", a.IdHandler)
	sm.HandleFunc("/diceware", a.DicewareHandler)

	// Front Page Handlers (related to auto gen from google docs)
	sm.HandleFunc("/page", a.FrontPageHandler)
	sm.HandleFunc("/refreshpage", a.RefreshPageHandler)

	// Root handler
	sm.HandleFunc("/", a.RootHandler)

	// Connect to couchdb server
	dbclient, err := kivik.New("couch", a.couchdbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database ", err)
	}
	a.db = dbclient.DB(context.TODO(), a.couchdbName)

	a.handler = c.Handler(sm)

	return &a
}

func (s *Server) Run() {
	s.FetchFrontPage()
	log.Println("Server listening at", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, s.handler))
}

func (s *Server) Login(user, pass string) (AuthDocument, string) {
	// selector
	q := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":     "auth",
			"username": user,
		},
	}

	rows, err := s.db.Find(context.TODO(), q)
	if err != nil {
		log.Println("Error retrieving login details", err)
		return AuthDocument{}, InternalServerError
	}

	if !rows.Next() { // Either no documents or an error
		err = rows.Err()
		if err != nil {
			log.Println("Error trying to access results ", err)
			return AuthDocument{}, InternalServerError
		} else {
			log.Println("No such user ", user)
			return AuthDocument{}, "no such user"
		}
	}

	var cur AuthDocument
	err = rows.ScanDoc(&cur)
	if err != nil {
		log.Println("Unable to unmarshal AuthDocument ", err)
		return AuthDocument{}, InternalServerError
	}

	// Now confirm password before handing back DB details

	newHash := hash(pass)

	if newHash != cur.Password { // Incorrect password
		log.Println("Incorrect login attempt for user", cur.Username)
		return AuthDocument{}, "incorrect password"
	}
	log.Println("Successful login attempt for ", cur.Username)
	return cur, ""
}

func (s *Server) FetchFrontPage() {
	// Check that we haven't refreshed
	d, _ := time.ParseDuration("1m")
	if time.Since(s.lastPageRefresh) < d {
		log.Println("Not refreshing, too soon")
		return
	}

	log.Println("Fetching front page...")
	ret, ok := getFrontPage(s.googleSheetCode)
	if ok != "" {
		log.Fatal(ok)
	}
	s.frontPage = ret
	s.lastPageRefresh = time.Now()
	log.Println("Finished fetching front page")
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	username, ok := query["user"]
	if !ok {
		w.Write(ErrHelper(UsernameNotProvided))
		return
	}

	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper(PasswordNotProvided))
		return
	}
	// We're good to go! Send them the details...

	ad, deets := s.Login(username[0], pass[0])
	if deets != "" { // Something happened
		w.Write(ErrHelper(deets))
		return
	}

	w.Write(LoginHelper(ad.Role, ad.UserID, s.dbAccessString))
}

func (s *Server) GenPendingAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	username, ok := query["user"]
	if !ok {
		w.Write(ErrHelper(UsernameNotProvided))
		return
	}

	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper(PasswordNotProvided))
		return
	}

	newRole, ok := query["role"]
	if !ok {
		w.Write(ErrHelper(RoleNotProvided))
		return
	}

	newUserID, ok := query["userid"]
	if !ok {
		w.Write(ErrHelper(UserIDNotProvided))
		return
	}

	ad, deets := s.Login(username[0], pass[0])

	if deets != "" { // Something happened
		w.Write(ErrHelper(deets))
		return
	}

	// Is this user an admin?

	if ad.Role != "Admin" {
		w.Write(ErrHelper(Unauthorized))
		return
	}

	pa := NewPendingAuthDocument(newUserID[0], newRole[0])
	// Put in DB

	_, err := s.db.Put(context.TODO(), pa.Id, pa)
	if err != nil {
		log.Println("Error saving pending auth ", err)
		w.Write(ErrHelper(InternalServerError))
		return
	}

	w.Write(SimpleHelper(pa.Code))

}

func (s *Server) CreateAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	username, ok := query["user"]
	if !ok {
		w.Write(ErrHelper(UsernameNotProvided))
		return
	}

	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper(PasswordNotProvided))
		return
	}

	code, ok := query["code"]
	if !ok {
		w.Write(ErrHelper(CodeNotProvided))
		return
	}

	usernameExists := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":     "auth",
			"username": username[0],
		},
	}
	rows, err := s.db.Find(context.TODO(), usernameExists)
	if err != nil {
		log.Println("Error retrieving existing usernames ", err)
		w.Write(ErrHelper(InternalServerError))
		return
	}
	if rows.Next() { // There is a currently matching document...
		log.Println("Can't create a duplicate with the same username", rows.Err())
		w.Write(ErrHelper(UsernameInUse))
		return
	}

	// selector
	q := map[string]interface{}{
		"selector": map[string]interface{}{
			"type": "pendingauth",
			"code": code[0],
		},
	}

	rows, err = s.db.Find(context.TODO(), q)
	if err != nil {
		log.Println("Error retrieving login details", err)
		w.Write(ErrHelper(InternalServerError))
		return
	}

	if !rows.Next() { // Either no documents or an error
		err = rows.Err()
		if err != nil {
			log.Println("Error trying to access results ", err)
			w.Write(ErrHelper(InternalServerError))
		} else {
			w.Write(ErrHelper(NoSuchPendingAuth))
		}
		return
	}

	var cur PendingAuthDocument
	err = rows.ScanDoc(&cur)
	if err != nil {
		log.Println("Unable to unmarshal PendingAuthDocument ", err)
		w.Write(ErrHelper(InternalServerError))
		return
	}

	// FIXME check for existing username

	nad := NewAuthDocument(cur.UserID, username[0], cur.Role, pass[0])

	_, err = s.db.Put(context.TODO(), nad.Id, nad)
	if err != nil {
		log.Println("Error creating new auth document ", err)
		w.Write(ErrHelper(InternalServerError))
		return
	}

	_, err = s.db.Delete(context.TODO(), cur.Id, cur.Rev)
	if err != nil {
		log.Println("Error deleting old pending auth ", err)
		w.Write(ErrHelper(InternalServerError))
	}

	w.Write(SimpleHelper(Success))
}

// Helper Handlers
func (s *Server) HashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper(PasswordNotProvided))
		return
	}

	w.Write(SimpleHelper(hash(pass[0])))
}

func (s *Server) TimestampHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(genTimestamp()))
}

func (s *Server) IdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(genID()))
}

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(WelcomeMessage))
}

func (s *Server) DicewareHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(genDiceware()))
}

func (s *Server) FrontPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ret, err := json.Marshal(s.frontPage)
	if err != nil {
		w.Write(ErrHelper(ErrorMarshalingResponse))
	} else {
		w.Write(SimpleHelper(string(ret)))
	}
}

func (s *Server) RefreshPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.FetchFrontPage()
	w.Write(SimpleHelper(Success))
}
