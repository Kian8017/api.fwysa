package main

const (
	UsernameNotProvided string = "no username provided"
	PasswordNotProvided string = "no password provided"
	RoleNotProvided     string = "no role provided"
	CodeNotProvided     string = "no code provided"
	UserIDNotProvided   string = "no userid provided"

	UsernameInUse     string = "username already in use"
	NoSuchPendingAuth string = "no such pending auth"

	Unauthorized        string = "unauthorized"
	InternalServerError string = "internal error"

	WelcomeMessage string = "welcome to the fwysa api server"
	Success        string = "success"

	// FrontPage specific
	ErrorFetchingPage       string = "error fetching page"
	ErrorReadingResponse    string = "error reading response"
	ErrorParsingPage        string = "error parsing page"
	ErrorParsingIndex       string = "error parsing an index (row/col)"
	ErrorMarshalingResponse string = "error marshaling response"

	// Upload specific
	ErrorGettingFile string = "error getting file from request"
)
