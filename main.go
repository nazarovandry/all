package main

import (
	"log"
	"net/http"
	"time"
	"io/ioutil"
	"sync"
	"strings"
	"strconv"
	"crypto/tls"

	"os"
	_ "github.com/heroku/x/hmetrics/onload"
)

//===[BASIC_FUNCTIONS]=======================================================\\

func site() (string) {
	return "https://elmacards.herokuapp.com/"
	//return "/"
}

func getCookies(r *http.Request) (*http.Cookie, bool) {
	session, err := r.Cookie("session_id")
	logged := err != http.ErrNoCookie
	if logged {
		return session, logged
	}
	return nil, logged
}

func adminName() (string) {
	return "AndrY"
}

func admin(logged bool, session *http.Cookie) (bool) {
	return logged && session.Value == adminName()
}

func wrong(s string) (bool) {
	return s == "" || strings.Contains(s, " ") || strings.Contains(s, "!") ||
		strings.Contains(s, "\n") || strings.Contains(s, "&") ||
		strings.Contains(s, "?") || strings.Contains(s, "\t") || len(s) > 30
}

func getShow (s string) string {
	if s == "on" {
		return "show"
	}
	return "hide"
}

func writeEnd(w http.ResponseWriter) {
	w.Write([]byte(`
			</div>
		</body>
		</html>
	`))
}

func hiddenPic() (string) {
	mu := &sync.Mutex{}
	mu.Lock()
	cards, _ := ioutil.ReadFile("cards.txt")
	mu.Unlock()
	array := strings.Split(string(cards), "\n")
	for _, ar := range array {
		first := strings.Split(string(ar), " ")
		if len(first) > 1 && first[0] == "hide" {
			return first[1]
		}
	}
	return ""
}

func howMany(a int, s string) (string) {
	if a == 1 {
		return s
	}
	return s + "s"
}

func openText(s string) (string) {
	if len(s) > 2 {
		return s[1 : len(s) - 2]
	}
	return ""
}

//===[WRITE_HTML_PAGE_BEGINNING]=============================================\\

func writeGeneral(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	w.Write([]byte(`<!doctype html>
	<html>
		<head>
			<title>Elma Cards Site</title>
			<style type="text/css">
			body {
				background:	#808080;
			}
			#head {
				background:	#DCDCDC;
				border:		1px groove black;
				padding:	10px;
			}
			.vertical {
				border-right:	1px solid black;
			}
			#text {
				background:	#DCDCDC;
				border:		1px groove black;
				width:		calc($(window).weight - 30px - $(#menu).width);
				padding:	10px;
				margin:		10px 0px 10px 10px;
				overflow:	scroll;
			}
			#menu {
				float:		left;
				background:	#DCDCDC;
				border:		1px groove black;
				width:		150px;
				padding:	10px;
				margin:		10px 10px 10px 0px;
			}
			#menu a {
				display:			block;
				color:				black;
				text-decoration:	none;
			}
			#text a {
				color:				blue;
				text-decoration:	none;
			}
			#text table, #text td, #text th {
				border:				1px solid black;
				border-collapse:	collapse;
				padding:			5px;
				vertical-align:		top;
			}
			#text th {
				background:	#29DD97;
			}
			</style>
		</head>
		<body>
			<div id="head">
				<table cellpadding="15">
				<tr>
				<td class="vertical">
					<p><font face="verdana" size="20"> Elma Cards </font></p>
				</td>
				<td>`))
	if logged {
		w.Write([]byte(`
		<form action="/action" method="post" class="reg-form">
		<div class="form-row">
			<p>Hi, ` + session.Value + `)</p>
			<p><input type="submit" name="action" value="Logout">
			<input type="submit" name="action" value="Change password">
			</p>
		</div>
		</form>`))
	} else {
		w.Write([]byte(`
		<form action="/login" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_name">Name: </label>
			<input type="text" id="form_name" name="name">
		</div>
		<div class="form-row">
			<label for="form_pw">Password: </label>
			<input type="password" id="form_pw" name="password">
		</div>
		<div class="form-row">
			<input type="submit" value="Oke">
		</div>
		</form>`))
	}
	w.Write([]byte(`	
		</td></tr></table>
		</div>
		<div id="menu">
			<div><a href="` + site() + `">Standings</a></div>
			<div><a href="` + site() + `contests">Contests</a></div>
			<div><a href="` + site() + `comments">Comments</a></div>
			<div><a href="` + site() + `cards">Cards</a></div>
			<p></p>
			<p><span style="color:#808080">&copy;AndrY 2019</span></p>
		</div>
		<div id="text">
	`))
}

//===[PAGES]=================================================================\\

func mainPage(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	writeGeneral(w, r)
	if admin(logged, session) {
		w.Write([]byte(`
		<form action="/users" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_name">Name: </label>
			<input type="text" id="form_name" name="name">
		</div>
		<div class="form-row">
			<label for="form_name">Value: </label>
    		<input type="text" id="form_value" name="password">
  		</div>
		<div class="form-row">
			<input type="submit" name="but" value="Add man">
			<input type="submit" name="but" value="Change password">
			<input type="submit" name="but" value="Change name">
			<input type="submit" name="but" value="Delete man">
		</div>
		</form>
		<form action="/addcard" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_cars">Card: </label>
			<input type="text" id="form_card" name="card">
		</div>
		<div class="form-row">
			<label for="form_name">For man: </label>
    		<input type="text" id="form_name" name="name">
			<label for="form_show">Shown: </label>
    		<input type="checkbox" id="form_show" name="shown">
  		</div>
		<div class="form-row">
			<input type="submit" value="Add card">
		</div>
		</form>
		<form action="/opercard" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_card">Card number: </label>
			<input type="text" id="form_card" name="num">
		</div>
		<div class="form-row">
			<label for="form_name">From man: </label>
    		<input type="text" id="form_name" name="name">
  		</div>
		<div class="form-row">
			<input type="submit" name="card_oper" value="Delete card">
			<input type="submit" name="card_oper" value="Make card shown">
			<input type="submit" name="card_oper" value="Make card hidden">
		</div>
		</form>
		<form action="/setpictures" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_name">Picture name: </label>
			<input type="text" id="form_name" name="name">
		</div>
		<div class="form-row">
			<label for="form_url">Url: </label>
    		<input type="text" id="form_name" name="url">
  		</div>
		<div class="form-row">
			<input type="submit" name="pic_oper" value="Delete pic">
			<input type="submit" name="pic_oper" value="Create/edit pic">
		</div>
		</form>
		<form action="/reload" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_saved">saved.txt </label>
			<textarea rows="3" cols="30" name="saved"></textarea>
		</div>
		<div class="form-row">
			<input type="submit" name="load" value="Reload">
		</div>
		</form>
		<form action="/download" method="post" class="reg-form">
		<div class="form-row">
			<input type="submit" name="load" value="Download">
		</div>
		</form>`))
	}
	w.Write([]byte(`
		</p>  Cards are given for participating in some cups, level packs` +
		` and etc. Probably these crads will be printed in plactic and sent` +
		` (ofc with better design).<p>`))
	if logged {
		w.Write([]byte(`
			<form action="/action" method="post" class="reg-form">
			<div class="form-row">
				<input type="submit" name="action" value="Show/hide cards">
			</div>
			</form>
		`))
	}
	w.Write([]byte(`
		<table bgcolor="white">
			<tr>
				<th>Name</th>
				<th>Cards</th>
			</tr>`))
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("list.txt")
	mu.Unlock()
	hiddenUrl := hiddenPic()
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) > 2 {
			w.Write([]byte(`<tr><td>` + elems[0] + `</td><td>`))
			links := strings.Split(elems[2], "&")
			for _, link := range links {
				pic := strings.Split(link, "?")
				if len(pic) > 1 && (pic[1] == "show" ||
					(logged && elems[0] == session.Value) ||
					admin(logged, session)){
					w.Write([]byte(`
					<img src=` + pic[0] + `>`))
				} else {
					w.Write([]byte(`
					<img src=` + hiddenUrl + `>`))
				}
			}
			w.Write([]byte(`</td></tr>`))
		}
	}
	w.Write([]byte(`</table>`))
	writeEnd(w)
}

func commPage(w http.ResponseWriter, r *http.Request) {
	_, logged := getCookies(r)
	writeGeneral(w, r)
	w.Write([]byte(`<form action="/send" method="post" class="reg-form">`))
	if !logged {
		w.Write([]byte(`
		<div class="form-row">
			<label for="form_url">Who are you? </label>
    		<input type="text" id="form_name" name="name">
  		</div>`))
	}
	w.Write([]byte(`
		<div class="form-row">
			<label for="form_list">Comment: </label>
			<textarea rows="1" cols="30" name="mess"></textarea>
			<input type="submit" name="send" value="Send">
		</div>
		</form>
		<p></p>`))
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("comm.txt")
	mu.Unlock()
	array := strings.Split(string(data), "\n")
	for _, ar := range array {
		first := strings.SplitN(string(ar), " ", 4)
		if len(first) >= 4 {
			w.Write([]byte(`<p><span style="color:#8B0000">[` +
				first[0] + ` ` + first[1] + `]</span> `))
			w.Write([]byte(`<b>` + first[2] + `: </b>` + first[3] + `</p>`))
		}
	}
	writeEnd(w)
}

func eventsPage(w http.ResponseWriter, r *http.Request) {
	writeGeneral(w, r)
	w.Write([]byte(`<p><b>Internals Inspired Cup (2019)</b> [ ` +
		`<a href="http://mopolauta.moposite.com/viewtopic.php?f` +
		`=3&p=264423#p=264423/">Info</a> | ` +
		`<a href="http://elmaonline.net/statistics/cups/13/">` +
		`Point standings</a> ]</p>`))
	writeEnd(w)
}

func cardsPage(w http.ResponseWriter, r *http.Request) {
	writeGeneral(w, r)
	w.Write([]byte(`
		<table border bgcolor="white">
			<tr>
				<th>Card</th>
				<th>Info</th>
			</tr>`))
	mu := &sync.Mutex{}
	mu.Lock()
	cards, _ := ioutil.ReadFile("cards.txt")
	list, _ := ioutil.ReadFile("list.txt")
	mu.Unlock()
	cardList := map[string]map[string]int{}
	infoList := map[string]string{}
	rows := strings.Split(string(cards), "\n")
	for _, row := range rows {
		elems := strings.SplitN(row, " ", 3)
		if len(elems) > 1 {
			cardList[elems[1]] = map[string]int{}
			if len(elems) > 2 {
				infoList[elems[1]] = openText(elems[2])
			}
		}
	}
	lines := strings.Split(string(list), "\n")
	for _, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) > 2 {
			links := strings.Split(elems[2], "&")
			for _, link := range links {
				pic := strings.Split(link, "?")
				if len(pic) > 1 && pic[1] == "show" {
					_, exists := cardList[pic[0]]
					if exists {
						_, exists2 := cardList[pic[0]][elems[0]]
						if exists2 {
							cardList[pic[0]][elems[0]] += 1
						} else {
							cardList[pic[0]][elems[0]] = 1
						}
					}
				}
			}
		}
	}
	for i, nameList := range cardList {
		w.Write([]byte(`<tr><td><img src=` + i + `></td>`))
		w.Write([]byte(`<td>`))
		total := 0
		info, exists := infoList[i]
		if exists {
			w.Write([]byte(`<p>` + info + `</p><hr>`))
		}
		for name, amount := range nameList {
			w.Write([]byte(`<p>` + name + `: ` +
				strconv.Itoa(amount) + ` ` + howMany(amount, "card") + `</p>`))
			total += amount
		}
		w.Write([]byte(`<p><b>Total amount is <span style="color:#DC143C">` +
			strconv.Itoa(total) + `</span></b></p>`))
		w.Write([]byte(`</td></tr>`))
	}
	w.Write([]byte(`</table>`))
	writeEnd(w)
}

//===[ADMIN_CARDS_OPERATIONS]================================================\\

func setPictures(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	if !logged || session.Value != "andry" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	name := r.FormValue("name")
	url := r.FormValue("url")
	button := r.FormValue("pic_oper")
	if wrong(name) || wrong(url) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("cards.txt")
	found := false
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) == 2 && elems[0] == name {
			if button == "Delete pic" {
				lines[i] = ""
			} else {
				found = true
				lines[i] = elems[0] + " " + url
			}
			continue
		}
	}
	if found == false && button == "Create/edit pic" {
		lines = append(lines, name + " " + url)
	}
	finlines := strings.Join(lines, "\n")
	finlines = strings.Replace(finlines, "\n\n", "\n", 1)
	if len(finlines) > 0 && finlines[0] == '\n' {
		finlines = finlines[1:]
	}
	_ = ioutil.WriteFile("cards.txt", []byte(finlines), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

func getUrl(card string) (string) {
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("cards.txt")
	mu.Unlock()
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) == 2 && elems[0] == card {
			return elems[1]
		}
	}
	return ""
}

//===[ADMIN_USERS_OPERATIONS]================================================\\

func users(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	name := r.FormValue("name")
	pass := r.FormValue("password")
	pass2 := r.FormValue("password2")
	button := r.FormValue("but")
	if wrong(pass) || (pass2 != "" && (wrong(pass2) || pass2 != pass)) ||
		button == "No, I dont want" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	log.Println(button)
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("list.txt")
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		elems := strings.Split(line, " ")
		if elems[0] == name ||
			(button == "Oke" && logged && session.Value == elems[0]) {
			if button == "Delete man" && pass == "delete" {
				lines[i] = ""
			} else if button == "Change name" && len(elems) > 1 {
				lines[i] = pass + " " + elems[1]
				if len(elems) > 2 {
					lines[i] += " " + elems[2]
				}
			} else if (button == "Change password" || button == "Oke") &&
				len(elems) > 1 {
				lines[i] = elems[0] + " " + pass
				if len(elems) > 2 {
					lines[i] += " " + elems[2]
				}
			} else if button == "Add man" {
				log.Println("Name is already exists.")
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			continue
		}
	}
	if button == "Add man" {
		lines = append(lines, name + " " + pass)
	}
	finlines := strings.Join(lines, "\n")
	finlines = strings.Replace(finlines, "\n\n", "\n", 1)
	if len(finlines) > 0 && finlines[0] == '\n' {
		finlines = finlines[1:]
	}
	_ = ioutil.WriteFile("list.txt", []byte(finlines), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

func operCard(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	name := r.FormValue("name")
	num := r.FormValue("num")
	button := r.FormValue("card_oper")
	number := -1
	if name != "" || button == "No, I dont want"{
		numberTmp, err := strconv.Atoi(num)
		if err != nil || button == "No, I dont want" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		number = numberTmp-1
	}
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("list.txt")
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) > 2 && (elems[0] == name ||
			(button == "Oke" && logged && session.Value == elems[0])) {
			newline := ""
			links := strings.Split(elems[2], "&")
			if len(links) > number || number == -1 {
				for j, link := range links {
					if j != number && number != -1 {
						newline += "&" + link
					} else if button == "Make card shown" {
						newline += "&" + strings.Split(link, "?")[0] + "?show"
					} else if button == "Make card hidden" {
						newline += "&" + strings.Split(link, "?")[0] + "?hide"
					} else if button == "Oke" {
						check := r.FormValue("shown" + strconv.Itoa(j))
						newline += "&" + strings.Split(link, "?")[0] +
							"?" + getShow(check)
					}
				}
				lines[i] = elems[0] + " " + elems[1]
				if len(newline) > 0 {
					lines[i] += " " + newline[1:]
				}
			}
		}
	}
	_ = ioutil.WriteFile("list.txt", []byte(strings.Join(lines, "\n")), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

func addCard(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	card := r.FormValue("card")
	name := r.FormValue("name")
	shown := r.FormValue("shown")
	url := getUrl(card)
	if url == "" || wrong(card) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("list.txt")
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		elems := strings.Split(line, " ")
		if len(elems) > 2 && elems[0] == name {
			lines[i] += "&" + getUrl(card) + "?" + getShow(shown)
		} else if len(elems) == 2 && elems[0] == name {
			lines[i] += " " + getUrl(card) + "?" + getShow(shown)
		}
	}
	_ = ioutil.WriteFile("list.txt", []byte(strings.Join(lines, "\n")), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[FULL_DATA]=============================================================\\

func download(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	if admin(logged, session) {
		mu := &sync.Mutex{}
		mu.Lock()
		list, _ := ioutil.ReadFile("list.txt")
		cards, _ := ioutil.ReadFile("cards.txt")
		comm, _ := ioutil.ReadFile("comm.txt")
		mu.Unlock()
		w.Write([]byte(list))
		w.Write([]byte("!"))
		w.Write([]byte(cards))
		w.Write([]byte("!"))
		w.Write([]byte(comm))
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func reload(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	saved := r.FormValue("saved")
	mu := &sync.Mutex{}
	mu.Lock()
	saved = strings.Replace(saved, "\r", "", -1)
	blocks := strings.SplitN(saved, "!", 3)
	if len(blocks) < 3 {
		return
		http.Redirect(w, r, "/", http.StatusFound)
	}
	_ = ioutil.WriteFile("list.txt", []byte(blocks[0]), 0644)
	_ = ioutil.WriteFile("cards.txt", []byte(blocks[1]), 0644)
	_ = ioutil.WriteFile("comm.txt", []byte(blocks[2]), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[LOGIN]=================================================================\\

func allRight(name string, pass string) (bool) {
	if wrong(name) || wrong(pass) {
		return false
	}
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("list.txt")
	mu.Unlock()
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		elem := strings.Split(l, " ")
		if len(elem) >= 2 && elem[0] == name {
			return elem[1] == pass
		}
	}
	return false
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(10 * time.Hour)
	name := r.FormValue("name")
	pass := r.FormValue("password")
	if allRight(name, pass) {
		cookie := http.Cookie{
			Name:    "session_id",
			Value:   name,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[USER_ACTIONS]==========================================================\\

func actionPage(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	button := r.FormValue("action")
	if button == "Logout" {
		if !logged {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else if button == "Show/hide cards" {
		writeGeneral(w, r)
		w.Write([]byte(`
		<form action="/opercard" method="post" class="reg-form">
		<table border bgcolor="white">
			<tr>
				<th>Card</th>
				<th>Shown</th>
			</tr>`))
		mu := &sync.Mutex{}
		mu.Lock()
		data, _ := ioutil.ReadFile("list.txt")
		mu.Unlock()
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			elems := strings.Split(line, " ")
			if len(elems) > 2 && logged && elems[0] == session.Value {
				links := strings.Split(elems[2], "&")
				for i, link := range links {
					pic := strings.Split(link, "?")
					if len(pic) > 1 {
						w.Write([]byte(`<tr>
							<td><img src=` + pic[0] + `></td>
							<td><div class="form-row">`))
						is := strconv.Itoa(i)
						if pic[1] == "hide" {
							w.Write([]byte(`
								<input type="checkbox" id="check` + is +
									 `" name="shown` + is + `">`))
						} else {
							w.Write([]byte(`
								<input type="checkbox" id="check` + is +
									`" name="shown` +
									is + `" checked="checked">`))
						}
						w.Write([]byte(`</div></td></tr>`))
					}
				}
			}
		}
		w.Write([]byte(`
		</table>
		<div class="form-row">
			<input type="submit" name="card_oper" value="Oke">
			<input type="submit" name="card_oper" value="No, I dont want">
		</div>
		</form>`))
		writeEnd(w)
		return
	}
	// else "Change Password"
	writeGeneral(w, r)
	w.Write([]byte(`
		<form action="/users" method="post" class="reg-form">
		<div class="form-row">
			<label for="form_passnew">New password: </label>
    		<input type="password" id="form_passnew" name="password">
  		</div>
		<div class="form-row">
			<label for="form_passnew2">New password again: </label>
    		<input type="password" id="form_passnew2" name="password2">
  		</div>
		<div class="form-row">
			<input type="submit" name="but" value="Oke">
			<input type="submit" name="but" value="No, I dont want">
		</div>
		</form>
	`))
	writeEnd(w)
}

func send(w http.ResponseWriter, r *http.Request) {
	session, logged := getCookies(r)
	name := r.FormValue("name")
	mess := r.FormValue("mess")
	send := r.FormValue("send")
	if send != "Send" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if !logged {
		name += "(?)"
	} else {
		name = session.Value
	}
	mu := &sync.Mutex{}
	mu.Lock()
	data, _ := ioutil.ReadFile("comm.txt")
	newdata :=  strings.Split(time.Now().String(), ".")[0] + " " +
		name + " " + mess + "\n" + string(data)
	_ = ioutil.WriteFile("comm.txt", []byte(newdata), 0644)
	mu.Unlock()
	http.Redirect(w, r, "/comments", http.StatusFound)
}

//===[BOT]===================================================================\\

func getBear(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!doctype html><html><body><p>TEST!</p></body></html>`))
	log.Println("frombot-Got")
}

func sendCat(w http.ResponseWriter, r *http.Request) {
	for {
		time.Sleep(2 * time.Minute)
		req, err := http.NewRequest(http.MethodDelete,
			"https://sdracamle.herokuapp.com/getbot", nil)
		if err == nil {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					},
				}
				client := &http.Client{
					Transport: tr,
					Timeout:   20 * time.Second,
				}
			_, err := client.Do(req)
			if err != nil {
				log.Println("client error: " + err.Error())
			} else {
				log.Println("tobot-Done")
			}
		} else {
			log.Println("request error" + err.Error())
		}
	}
}

//===[MAIN]==================================================================\\

func main() {
	list, _ := ioutil.ReadFile("list.txt")
	cards, _ := ioutil.ReadFile("cards.txt")
	comm, _ := ioutil.ReadFile("comm.txt")
	_ = ioutil.WriteFile("list.txt",
		[]byte(strings.Replace(string(list), "\r", "", -1)), 0644)
	_ = ioutil.WriteFile("cards.txt",
		[]byte(strings.Replace(string(cards), "\r", "", -1)), 0644)
	_ = ioutil.WriteFile("comm.txt",
		[]byte(strings.Replace(string(comm), "\r", "", -1)), 0644)

	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/action", actionPage)
	http.HandleFunc("/addcard", addCard)
	http.HandleFunc("/opercard", operCard)
	http.HandleFunc("/users", users)
	http.HandleFunc("/setpictures", setPictures)
	http.HandleFunc("/reload", reload)
	http.HandleFunc("/download", download)
	http.HandleFunc("/comments", commPage)
	http.HandleFunc("/send", send)
	http.HandleFunc("/contests", eventsPage)
	http.HandleFunc("/cards", cardsPage)
	http.HandleFunc("/sendbot", sendCat)
	http.HandleFunc("/getbot", getBear)
	http.HandleFunc("/", mainPage)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	http.ListenAndServe(":"+port, nil)

	log.Println("starting server at :8080")
	//http.ListenAndServe(":8080", nil)
}
