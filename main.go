package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var secret string
var router *gin.Engine
var store cookie.Store

func main() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/event_management_solution") // Connecting to database
	if err != nil {
		log.Println("ERROR: ", err)
	}
	defer db.Close()

	gob.Register(User{}) // This method is used to create a "binary template" in storage so we can store our user inside a session

	router = gin.Default()

	secret = "kuQm6IyJK8hMN/ZxQw5cUNWE7nzagQ/CHbqmXNP+qmldNvtqNeko/J+S/pJB1Lwb" // Secret key which is used to access session storage

	store = cookie.NewStore([]byte(secret))
	router.Use(sessions.Sessions("usersession", store))

	router.Static("/static", "./web")
	router.LoadHTMLGlob("web/html/*.html") // Enables access to static files such as html, css and javscript in this project

	// WEB: (These requests return html files which are loaded in clients browser)
	router.GET("/login", handleLogin)
	router.POST("/login", handleLoginSubmission)
	router.GET("/register", handleRegister)
	router.POST("/register", handleRegisterSubmission)
	router.GET("/logout", handleLogout)
	router.GET("/", handleHome)
	router.GET("/organizations", handleOrganizations)
	router.GET("/events", handleEvents)
	router.GET("/meetings", handleMeetings)
	router.GET("/myInvitations", handleInvitationInbox)
	router.GET("/mySchedule", handleSchedule)

	// JSON REST API end-points:
	router.GET("/getOrganizations", handleGetOrganizations)
	router.POST("/addOrganization", handleAddOrganization)
	router.GET("/getUserByID", handleGetUserByID)
	router.GET("/getEvents", handleGetEvents)
	router.POST("/addEvent", handleAddEvent)
	router.POST("/joinEvent", handleJoinEvent)
	router.POST("/leaveEvent", handleLeaveEvent)
	router.POST("/getAttendees", handleGetAttendees)
	router.GET("/getParticipations", handleGetParticipations)
	router.POST("/sendInvitations", handleSendInvitations)
	router.POST("/getMeeting", handleGetMeeting)
	router.GET("/getUsersMeetings", handleGetUsersMeetings)
	router.POST("/getInviteesForMeeting", handleGetInviteesForMeeting)
	router.GET("/getUsersInvitations", handleGetUserInvitations)
	router.POST("/acceptInvitation", handleAcceptInvitation)
	router.POST("/rejectInvitation", handleRejectInvitation)
	router.POST("/getPossibleTimes", handleGetPossibleTimes)
	router.POST("/scheduleMeetingTime", handleScheduleMeetingTime)
	router.GET("/getUserSchedule", handleGetUserSchedule)

	log.Fatal(router.Run(":8080"))
}

// HANDLERS

func handleHome(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		user := session.Get("userProfile").(User)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"username": user.Username,
			"email":    user.Email,
		})
	} else {
		log.Println("Authentication status:", authenticated)
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func handleLoginSubmission(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	usernameAuthData, err := getUserAuthData(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	log.Println("SUCCESS! Found AuthData for: ", username)

	if strings.Compare(password, usernameAuthData.Password) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID PASSWORD"})
		return
	}
	log.Println("SUCCESS! Passwords matching for: ", username)

	userProfile, err := getUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errorGetProfile": err})
		return
	}
	log.Println("SUCCESS! Found profile for: ", username)

	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Set("userProfile", userProfile)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		log.Println(err)
		return
	}

	log.Println("Session after login:", session.Get("authenticated"))

	c.Redirect(http.StatusFound, "/")
}

func handleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.Redirect(http.StatusSeeOther, "/login")
}

func handleRegister(c *gin.Context) {
	organizations, err := getOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"Organizations": organizations,
	})
}

func handleRegisterSubmission(c *gin.Context) {
	// Podaci se trenutno ne provjeravaju ni u front endu ni u back endu, potrebno je dodati provjere poput REGEX-a
	// kako bi svi podaci koji se šalju/primaju bili validni

	username := c.PostForm("username")
	password := c.PostForm("password")
	email := c.PostForm("email")
	organizationId := c.PostForm("organization")

	organizationIdInt, err := strconv.Atoi(organizationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	usernameUsed, err := checkUsernameExists(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if usernameUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "USERNAME TAKEN"})
		return
	}

	activeUser, err := createUser(User{0, username, email, organizationIdInt}, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Set("activeUser", activeUser)
	session.Save()

	c.Redirect(http.StatusFound, "/login")
}

func handleOrganizations(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		c.HTML(http.StatusOK, "organizations.html", nil)
	} else {
		log.Println("Authentication status:", authenticated)
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleEvents(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		c.HTML(http.StatusOK, "eventsAuthenticated.html", nil)
	} else {
		c.HTML(http.StatusOK, "eventsUnauthenticated.html", nil)
	}
}

func handleMeetings(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		c.HTML(http.StatusOK, "meetings.html", nil)
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleInvitationInbox(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		c.HTML(http.StatusOK, "invitationInbox.html", nil)
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleSchedule(c *gin.Context) {
	session := sessions.Default(c)
	authenticated := session.Get("authenticated")
	if authenticated != nil && authenticated.(bool) {
		c.HTML(http.StatusOK, "schedule.html", nil)
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

// JSON REST API

// All of these functions work similarly, they either handle POST or GET requests. Ones which handle POST request parse data from the
// request body which is in JSON format to one of the structs which are defined at the end of this document. GET requests use session
// data if needed and return data from the database. Both types of methods return a response with either an error status
// (e.g. BAD REQUEST, INTERNAL SERVER ERROR) or a success status (e.g. OK)

func handleAddOrganization(c *gin.Context) {
	var org Organization

	if err := c.BindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := insertOrganization(&org)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

func handleGetOrganizations(c *gin.Context) {
	organizations, err := getOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, organizations)
}

func handleAddEvent(c *gin.Context) {
	var event Event

	if err := c.BindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := insertEvent(&event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func handleGetEvents(c *gin.Context) {
	events, err := getEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func handleJoinEvent(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participationId, err := insertParticipation(user.UserID, request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Participation registered successfully!",
		"participationId": participationId,
	})
}

func handleLeaveEvent(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = removeParticipation(user.UserID, request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed participation successfully!"})
}

func handleGetParticipations(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	participations, err := getUserParticipations(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participations)
}

func handleGetAttendees(c *gin.Context) {
	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attendees, err := getAttendees(request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("GOT THEM USERS! SENDING BACK!")
	log.Println(len(attendees))

	c.JSON(http.StatusOK, attendees)
}

func handleSendInvitations(c *gin.Context) {
	var request CreateMeetingRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	registeredId, err := createMeeting(request.SelectedEvent, request.Duration, user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, invitee := range request.Invitees {
		err := createInvitation(registeredId, invitee.EventID)
		if err != nil {
			log.Println("Error occured: ", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"meetingId": registeredId,
	})
}

func handleGetMeeting(c *gin.Context) {
	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meeting, err := getMeetingById(request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, meeting)
}

func handleGetUsersMeetings(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	meetings, err := getMeetingsOf(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println(meetings)

	c.JSON(http.StatusOK, meetings)
}

func handleGetInviteesForMeeting(c *gin.Context) {
	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteesAndStatus, err := getInviteesAndStatus(request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inviteesAndStatus)
}

func handleGetUserInvitations(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	invitations, err := getInvitationsFor(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

func handleGetUserByID(c *gin.Context) {
	userId, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := getUserById(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func handleAcceptInvitation(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = changeInvitationStatus(user.UserID, request.EventID, "Accepted")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully accepted the invitation!"})
}

func handleRejectInvitation(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = changeInvitationStatus(user.UserID, request.EventID, "Rejected")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully rejected the invitation!"})
}

func handleGetPossibleTimes(c *gin.Context) {
	var request EventRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	possible, err := canScheduleTime(request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !possible {
		data := []string{}
		data = append(data, "Not all users have decided!")
		c.JSON(http.StatusOK, TimeOptions{"unavailable", data})
		return
	}

	possibleTimes, err := calculatePossibleTimes(request.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TimeOptions{"available", possibleTimes})
}

func handleScheduleMeetingTime(c *gin.Context) {
	var request ScheduleRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meeting, err := getMeetingById(request.MeetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	invitees, err := getInviteesAndStatus(meeting.MeetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endTime, err := calculateEndTime(request.Time, meeting.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, invitee := range invitees {
		if invitee.InvitationInfo.Status != "Accepted" {
			continue
		}

		scheduleEntry := ScheduleEntry{invitee.Invitee.UserID, meeting.EventID, request.Time, endTime, meeting.MeetingID}

		err := updateUserSchedule(scheduleEntry)
		if err != nil {
			log.Println("ERROR: ", err.Error())
			continue
		}
	}

	scheduleEntry := ScheduleEntry{meeting.OrganizatorID, meeting.EventID, request.Time, endTime, meeting.MeetingID}

	err = updateUserSchedule(scheduleEntry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = updateMeetingScheduledTime(meeting.MeetingID, request.Time)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Meeting scheduled successfully!"})
}

func handleGetUserSchedule(c *gin.Context) {
	auth, user := checkAuthAndGetUser(c)
	if !auth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied, you must authenticate!"})
		return
	}

	if user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, try to authenticating again."})
		return
	}

	schedule, err := getUserSchedule(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DATABASE

// These methods are uset to handle the work with MySQL database. They have SQL statements and are taking in needed paramters with which
// they use to get data from database, update data in database or insert data into database. They return error (and last insert ID sometimes)
// so that we know what kind of responds we will send and how we will handle the situation

func insertOrganization(org *Organization) error {
	insertQuery := "INSERT INTO organizations (organization_name) VALUES (?)"
	_, err := db.Exec(insertQuery, org.OrganizationName)
	if err != nil {
		return err
	}

	return nil
}

func insertEvent(event *Event) error {
	insertQuery := "INSERT INTO events (name, date, organization_id, start_time, end_time) VALUES (?, ?, ?, ?, ?)"

	_, err := db.Exec(insertQuery, event.Name, event.Date, event.OrganizatorID, event.StartTime, event.EndTime)
	if err != nil {
		return err
	}

	return nil
}

func getUserAuthData(username string) (AuthData, error) {
	selectQuery := "SELECT * FROM authentication WHERE username = ?"

	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		log.Println("Failed to prepare SELECT statement!")
		return AuthData{}, err
	}
	defer stmt.Close()

	var aData AuthData
	err = stmt.QueryRow(username).Scan(&aData.UserID, &aData.Username, &aData.Password, &aData.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user fount with username:", username)
			return AuthData{}, nil
		}
		log.Println("Failed to execute the SELECT query:", err)
		return AuthData{}, err
	}

	return aData, nil
}

func checkUsernameExists(username string) (bool, error) {
	selectQuery := "SELECT COUNT(*) FROM authentication WHERE username = ?"

	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		log.Println("Failed to prepare SELECT statement!")
		return true, err
	}
	defer stmt.Close()

	var count int

	err = db.QueryRow(selectQuery, username).Scan(&count)
	if err != nil {
		log.Println("Couldn't execute query! Error:", err)
		return true, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

func createUser(user User, password string) (User, error) {
	insertQuery := "INSERT INTO users (name, email, organization_id) VALUES (?, ?, ?)"
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Username, user.Email, user.OrganizationID)
	if err != nil {
		return User{}, err
	}

	usersId, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	insertQuery = "INSERT INTO authentication (user_id, username, password, auth_token) VALUES (?, ?, ?, ?)"
	stmt, err = db.Prepare(insertQuery)
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(usersId, user.Username, password, secret)
	if err != nil {
		return User{}, err
	}

	return User{int(usersId), user.Username, user.Email, user.OrganizationID}, nil
}

func getOrganizations() ([]Organization, error) {
	selectQuery := "SELECT organization_id, organization_name FROM organizations"
	rows, err := db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	organizations := []Organization{}
	for rows.Next() {
		var organization Organization
		err := rows.Scan(&organization.ID, &organization.OrganizationName)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}

	return organizations, nil
}

func getEvents() ([]Event, error) {
	selectQuery := "SELECT * FROM events"
	rows, err := db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.EventID, &event.Name, &event.Date, &event.OrganizatorID, &event.StartTime, &event.EndTime)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func getUser(username string) (*User, error) {
	selectQuery := "SELECT user_id, name, email, organization_id FROM users WHERE name = ?"

	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := &User{}
	err = stmt.QueryRow(username).Scan(&user.UserID, &user.Username, &user.Email, &user.OrganizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found!")
			return nil, nil
		}
		log.Println(err)
		return nil, err
	}

	log.Println("Got user from db!")
	return user, nil
}

func getUserById(userId int) (*User, error) {
	selectQuery := "SELECT user_id, name, email, organization_id FROM users WHERE user_id = ?"

	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := &User{}
	err = stmt.QueryRow(userId).Scan(&user.UserID, &user.Username, &user.Email, &user.OrganizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Println(err)
		return nil, err
	}

	return user, nil
}

func getEventById(eventId int) (*Event, error) {
	selectQuery := "SELECT * FROM events WHERE event_id = ?"
	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	event := &Event{}
	err = stmt.QueryRow(eventId).Scan(&event.EventID, &event.Name, &event.Date, &event.OrganizatorID, &event.StartTime, &event.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Println(err)
		return nil, err
	}

	return event, nil
}

func insertParticipation(userId int, eventId int) (int, error) {
	insertQuery := "INSERT INTO user_events (user_id, event_id) VALUES (?, ?)"

	res, err := db.Exec(insertQuery, userId, eventId)
	if err != nil {
		return 0, err
	}

	participationId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(participationId), nil
}

func removeParticipation(userId int, eventId int) error {
	deleteQuery := "DELETE FROM user_events WHERE user_id = ? AND event_id = ?"

	_, err := db.Exec(deleteQuery, userId, eventId)
	if err != nil {
		return err
	}

	return nil
}

func getUserParticipations(id int) ([]Event, error) {
	query := `
		SELECT events.event_id, events.name, events.date, events.organization_id, events.start_time, events.end_time
		FROM user_events
		JOIN events ON user_events.event_id = events.event_id
		WHERE user_events.user_id = ?
	`
	// Execute the query
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.EventID, &event.Name, &event.Date, &event.OrganizatorID, &event.StartTime, &event.EndTime)
		if err != nil {
			log.Println("Error scanning event row:", err)
			continue
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func getAttendees(eventId int) ([]User, error) {
	query := `
		SELECT users.user_id, users.name, users.email, users.organization_id
		FROM user_events
		JOIN users ON user_events.user_id = users.user_id
		WHERE user_events.event_id = ?;
	`

	rows, err := db.Query(query, eventId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.OrganizationID)
		if err != nil {
			log.Println("Error scanning user row:", err.Error())
			continue
		}

		users = append(users, user)

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return users, nil
}

func createMeeting(selectedEvent Event, duration int, currentUserId int) (int, error) {
	insertQuery := "INSERT INTO meetings (event_id, scheduled_date, scheduled_time, organizer_id, duration) VALUES (?, ?, ?, ?, ?)"
	res, err := db.Exec(insertQuery, selectedEvent.EventID, selectedEvent.Date, "", currentUserId, duration)
	if err != nil {
		return 0, err
	}

	meetingId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = createMeetingSchedulingStatus(int(meetingId), currentUserId)
	if err != nil {
		return 0, err
	}

	return int(meetingId), nil
}

func createInvitation(meetingId int, inviteeId int) error {
	insertQuery := "INSERT INTO meeting_invitees (meeting_id, invitee_id, status) VALUES (?, ?, ?)"

	_, err := db.Exec(insertQuery, meetingId, inviteeId, "Pending")
	if err != nil {
		return err
	}

	return nil
}

func getMeetingsOf(userId int) ([]Meeting, error) {
	log.Println("Gettings meetings for:", userId)
	selectQuery := "SELECT meeting_id, event_id, scheduled_date, scheduled_time, organizer_id, duration FROM meetings WHERE organizer_id = ?"

	rows, err := db.Query(selectQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	meetings := []Meeting{}
	for rows.Next() {
		var meeting Meeting
		err := rows.Scan(&meeting.MeetingID, &meeting.EventID, &meeting.Date, &meeting.Time, &meeting.OrganizatorID, &meeting.Duration)
		if err != nil {
			log.Println("Error scanning user row:", err.Error())
			continue
		}

		meetings = append(meetings, meeting)

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return meetings, nil
}

func getInviteesAndStatus(meetingId int) ([]InviteeAndStatus, error) {
	invitationSelectQuery := "SELECT meeting_id, invitee_id, status FROM meeting_invitees WHERE meeting_id = ?"

	rows, err := db.Query(invitationSelectQuery, meetingId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invitations := []Invitation{}

	for rows.Next() {
		var invitation Invitation
		err := rows.Scan(&invitation.MeetingID, &invitation.InviteeID, &invitation.Status)
		if err != nil {
			log.Println("Error scanning user row:", err.Error())
			continue
		}

		invitations = append(invitations, invitation)

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	resultArray := []InviteeAndStatus{}

	for _, invitation := range invitations {
		user, err := getUserById(invitation.InviteeID)
		if err != nil {
			log.Println("Error occured fetching user with ID:", invitation.InviteeID)
			continue
		}

		entry := InviteeAndStatus{*user, invitation}
		resultArray = append(resultArray, entry)
	}

	return resultArray, nil
}

func getInvitationsFor(userId int) ([]Invitation, error) {
	selectQuery := "SELECT meeting_id, invitee_id, status FROM meeting_invitees WHERE invitee_id = ?"

	rows, err := db.Query(selectQuery, userId)
	if err != nil {
		return nil, err
	}

	invitations := []Invitation{}
	for rows.Next() {
		var invitation Invitation
		err := rows.Scan(&invitation.MeetingID, &invitation.InviteeID, &invitation.Status)
		if err != nil {
			log.Println("Error occured fetching invitation: ", err)
			continue
		}
		invitations = append(invitations, invitation)

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return invitations, nil
}

func getMeetingById(meetingId int) (*Meeting, error) {
	selectQuery := "SELECT meeting_id, event_id, scheduled_date, scheduled_time, organizer_id, duration FROM meetings WHERE meeting_id = ?"

	stmt, err := db.Prepare(selectQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	meeting := &Meeting{}
	err = stmt.QueryRow(meetingId).Scan(&meeting.MeetingID, &meeting.EventID, &meeting.Date, &meeting.Time, &meeting.OrganizatorID, &meeting.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return meeting, nil
}

func changeInvitationStatus(userId int, meetingId int, newStatus string) error {
	updateQuery := "UPDATE meeting_invitees SET status = ? WHERE meeting_id = ? AND invitee_id = ?"

	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newStatus, meetingId, userId)
	if err != nil {
		return err
	}

	err = calculateScheduledTimeStatus(meetingId)
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}

func createMeetingSchedulingStatus(meetingId int, organizerId int) error {
	insertQuery := "INSERT INTO meeting_scheduling_status (organizer_id, meeting_id, status) VALUES (?, ?, ?)"

	_, err := db.Exec(insertQuery, organizerId, meetingId, "Pending")
	if err != nil {
		return err
	}

	return nil
}

func updateUserSchedule(entry ScheduleEntry) error {
	insertQuery := "INSERT INTO user_schedule (user_id, event_id, start_time, end_time, meeting_id) VALUES (?, ?, ?, ?, ?)"

	_, err := db.Exec(insertQuery, entry.UserID, entry.EventID, entry.StartTime, entry.EndTime, entry.MeetingID)
	if err != nil {
		log.Println("ERROR INSERTING INTO SCHEDULE: ", err.Error())
		return err
	}

	return nil
}

func updateMeetingScheduledTime(meetingId int, time string) error {
	updateQuery := "UPDATE meetings SET scheduled_time = ? WHERE meeting_id = ?"

	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time, meetingId)
	if err != nil {
		return err
	}

	return nil
}

// OPERATIONAL

// checkAuthAndGetUser

// this method checks if user is authenticated and then tries to access his profile information from the session, it returns boolean
// that tells us if the user is authenticated or isn't (if he is returns true, if not returns false) and it returns a pointer to an
// object which represents users profile
func checkAuthAndGetUser(c *gin.Context) (bool, *User) {
	session := sessions.Default(c)

	authenticated := session.Get("authenticated")
	if authenticated == nil || !authenticated.(bool) {
		return false, nil
	}

	user := session.Get("userProfile").(User)
	empty := User{}
	if user == empty {
		return false, nil
	}

	return true, &user
}

// calculateScheduledTimeStatus

// this method gets all invitation statuses for a specific meeting, and checks if all users have determined if they are coming or not
// if all have determined they update a row in meeting_scheduling_status table so that we can know can we allow meeting organizer to
// pick a time for the meeting or not
func calculateScheduledTimeStatus(meetingId int) error {
	selectInviteesQuery := "SELECT meeting_id, invitee_id, status FROM meeting_invitees WHERE meeting_id = ?"

	rows, err := db.Query(selectInviteesQuery, meetingId)
	if err != nil {
		return err
	}

	for rows.Next() {
		var invitation Invitation
		err := rows.Scan(&invitation.MeetingID, &invitation.InviteeID, &invitation.Status)
		if err != nil {
			log.Println("Error occured fetching invitation: ", err)
			continue
		}

		if invitation.Status == "Pending" {
			return nil
		}

		if err := rows.Err(); err != nil {
			return err
		}
	}

	updateQuery := "UPDATE meeting_scheduling_status SET status = ? WHERE meeting_id = ?"

	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec("Ready", meetingId)
	if err != nil {
		return err
	}

	return nil
}

// canScheduleTime

// this method is connected to the last one it just tells if can we allow user to pick a time or not
func canScheduleTime(meetingId int) (bool, error) {
	var schStatus string
	selectQuery := "SELECT status FROM meeting_scheduling_status WHERE meeting_id = ?"
	err := db.QueryRow(selectQuery, meetingId).Scan(&schStatus)
	if err != nil {
		log.Println("EROOR: ", err.Error())
		return false, err
	}

	if schStatus == "Ready" {
		return true, nil
	} else {
		return false, nil
	}

}

// parseTime

// as we access time as a string in format "hh:mm:ss" this function allows us to parse it into time.Time type
func parseTime(timeStr string) (time.Time, error) {
	layout := "15:04:05"

	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

func getUserSchedule(userId int) (*Schedule, error) {
	selectQuery := "SELECT * FROM user_schedule WHERE user_id = ?"

	rows, err := db.Query(selectQuery, userId)
	if err != nil {
		return nil, err
	}

	var schedule Schedule

	for rows.Next() {
		var scheduleEntry ScheduleEntry
		err = rows.Scan(&scheduleEntry.UserID, &scheduleEntry.EventID, &scheduleEntry.StartTime, &scheduleEntry.EndTime, &scheduleEntry.MeetingID)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		schedule.ScheduleEntries = append(schedule.ScheduleEntries, scheduleEntry)

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return &schedule, nil
}

// filterSchedule

// takes in a schedule from an user and an event ID and creates schedule for that specific event

func filterSchedule(schedule Schedule, meetingId int) (*Schedule, error) {
	meeting, err := getMeetingById(meetingId)
	if err != nil {
		return nil, err
	}

	filteredSchedule := Schedule{}

	for _, scheduleEntry := range schedule.ScheduleEntries {
		entryMeeting, err := getEventById(scheduleEntry.EventID)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if entryMeeting.Date == meeting.Date {
			filteredSchedule.ScheduleEntries = append(filteredSchedule.ScheduleEntries, scheduleEntry)
		}
	}

	return &filteredSchedule, nil
}

// getInviteesSchedule:

// returns an array of schedules for all users that are coming to the meeting
func getInviteesSchedules(meetingId int) ([]Schedule, error) {
	invitees, err := getInviteesAndStatus(meetingId)
	if err != nil {
		return nil, err
	}

	schedules := []Schedule{}

	for _, invitee := range invitees {
		if invitee.InvitationInfo.Status != "Accepted" {
			continue
		}

		schedule, err := getUserSchedule(invitee.Invitee.UserID)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		schedule, err = filterSchedule(*schedule, meetingId)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		schedules = append(schedules, *schedule)
	}

	return schedules, nil
}

func calculatePossibleTimes(meetingId int) ([]string, error) {
	meeting, err := getMeetingById(meetingId)
	if err != nil {
		return nil, err
	}

	event, err := getEventById(meeting.EventID)
	if err != nil {
		return nil, err
	}

	userSchedules, err := getInviteesSchedules(meetingId)
	if err != nil {
		return nil, err
	}

	var possibleTimes []string

	startTime, err := parseTime(event.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := parseTime(event.EndTime)
	if err != nil {
		return nil, err
	}

	currentTime := startTime

	// Iterate through each minute from the start time to the end time and check if there is any collision for any of the invitees
	// if there is we consider that time unavailable and move on
	for currentTime.Before(endTime) {
		available := true

		for _, schedule := range userSchedules {
			for _, entry := range schedule.ScheduleEntries {
				entryStartTime, err := parseTime(entry.StartTime)
				if err != nil {
					return nil, err
				}
				entryEndTime, err := parseTime(entry.EndTime)
				if err != nil {
					return nil, err
				}

				if currentTime.After(entryStartTime) && currentTime.Before(entryEndTime) {
					available = false
					break
				}
			}

			if !available {
				break
			}
		}

		if available {
			possibleTimes = append(possibleTimes, currentTime.Format("15:04:05"))
		}

		currentTime = currentTime.Add(time.Minute)
	}

	return possibleTimes, nil
}

// calculateEndTime:

// this method takes in a startTime as string in format "hh:mm:ss" and a durtionInMinutes integer and then gives us what is the
// end time of a meeting
func calculateEndTime(startTime string, durationInMinutes int) (string, error) {
	layout := "15:04:05"

	startTimeObj, err := time.Parse(layout, startTime)
	if err != nil {
		return "", err
	}

	duration := time.Duration(durationInMinutes) * time.Minute

	endTimeObj := startTimeObj.Add(duration)

	endTime := endTimeObj.Format(layout)

	return endTime, nil
}

// MODELS

type Organization struct {
	ID               string `json:"id"`
	OrganizationName string `json:"organization_name"`
}

type AuthData struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"auth_token"`
}

type User struct {
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	OrganizationID int    `json:"organization_id"`
}

type Event struct {
	EventID       int    `json:"id"`
	Name          string `json:"name"`
	Date          string `json:"date"`
	OrganizatorID string `json:"organizator"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
}

type EventRequest struct {
	EventID int `json:"id"`
}

type EventParticipation struct {
	ParticipationID int `json:"participation_id"`
	UserID          int `json:"user_id"`
	EventID         int `json:"event_id"`
}

type CreateMeetingRequest struct {
	SelectedEvent Event          `json:"event"`
	Duration      int            `json:"duration"`
	Invitees      []EventRequest `json:"invited_users_ids"`
}

type Meeting struct {
	MeetingID     int    `json:"meeting_id"`
	EventID       int    `json:"event_id"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	OrganizatorID int    `json:"organizator_id"`
	Duration      int    `json:"duration"`
}

type InviteeAndStatus struct {
	Invitee        User       `json:"invitee"`
	InvitationInfo Invitation `json:"invitation"`
}

type Invitation struct {
	MeetingID int    `json:"meeting_id"`
	InviteeID int    `json:"invitee_id"`
	Status    string `json:"status"`
}

type TimeOptions struct {
	Status string   `json:"status"`
	Times  []string `json:"time"`
}

type ScheduleEntry struct {
	UserID    int    `json:"user_id"`
	EventID   int    `json:"event_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	MeetingID int    `json:"meeting_id"`
}

type Schedule struct {
	ScheduleEntries []ScheduleEntry `json:"schedule_entries"`
}

type ScheduleRequest struct {
	MeetingID int    `json:"meeting_id"`
	Time      string `json:"time"`
}
