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

//===[STRUCTURES]============================================================\\

type Card struct {
	url			string
	info		string
}

type MyCard struct {
	name		string
	show		bool
	info		string
}

type Man struct {
	password	string
	cards		[]*MyCard
}

type Comm struct {
	name		string
	text		string
	time		string
}

type All struct {
	cards		map[string]*Card
	mans		map[string]*Man
	comms		[]*Comm
	mu			sync.Mutex
	mainPage	string
	eventsPage	string
	bot			*int
}

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
	return s == "" || strings.Contains(s, "\t") ||
		strings.Contains(s, "\n") || strings.Contains(s, "(-BLOCK-)") ||
		strings.Contains(s, "(-STRING-)") ||
		strings.Contains(s, "(-ELEM-)") ||
		strings.Contains(s, "(-THING-)") ||
		strings.Contains(s, "(-PART-)") ||
		strings.ContainsAny(s, "<>/\\'{}`\"") || len(s) > 30
}

func checkText(s *string) {
	*s = strings.Replace(*s, "\r", "\n", -1)
	*s = strings.Replace(*s, "\r", "\n", -1)
	*s = strings.Replace(*s, "(-BLOCK-)", "", -1)
	*s = strings.Replace(*s, "(-STRING-)", "", -1)
	*s = strings.Replace(*s, "(-ELEM-)", "", -1)
	*s = strings.Replace(*s, "(-THING-)", "", -1)
	*s = strings.Replace(*s, "(-PART-)", "", -1)
	if len(*s) > 500 {
		*s = (*s)[0:500]
	}
}

func code(input string) (string){
	s := input
	set := "<>/\\'{}`\""
	for _, i := range set {
		s = strings.Replace(s, string(i),
			"&#" + strconv.Itoa(int(i)) + ";", -1)
	}
	return s
}

func isOn(s string) bool {
	return s == "on"
}

func isShow(s string) bool {
	return s == "show"
}

func isTrue(b bool) string {
	if b == true {
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

func hiddenPic(cards map[string]*Card) (*Card) {
	hide, exists := cards["hide"]
	if !exists {
		card := Card {
			url:	"hidden",
			info:	"bad url",
		}
		cards["hide"] = &card
		return &card
	}
	return hide
}

func howMany(a int, s string) (string) {
	if a == 1 {
		return s
	}
	return s + "s"
}

//===[WRITE_HTML_PAGE_BEGINNING]=============================================\\

func writeGeneral(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	w.Write([]byte(`<!doctype html>
	<html>
		<head>
			<title>Elma Cards Site</title>
			<style type="text/css">
			p {
				padding-top:	0px;
				line-height:	1.5;
				font-family:	Verdana, Geneva, sans-serif;
			}
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
				padding:			10px;
				vertical-align:		top;
				text-align:			left;
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
		<div class="form-row"><p>
			<label for="form_name">Name: </label>
			<input type="text" id="form_name" name="name"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_pw">Password: </label>
			<input type="password" id="form_pw" name="password">
			<input type="submit" value="Oke"></p>
		</div>
		</form>`))
	}
	w.Write([]byte(`	
		</td></tr></table>
		</div>
		<div id="menu">
			<div><p><a href="` + site() + `">Standings</a></p></div>
			<div><p><a href="` + site() + `contests">Contests</a></p></div>
			<div><p><a href="` + site() + `comments">Comments</a></p></div>
			<div><p><a href="` + site() + `cards">Cards</a></div>
			<p></p>
			<p><span style="color:#808080">&copy;AndrY 2019</span></p>
		</div>
		<div id="text">
	`))
}

//===[PAGES]=================================================================\\

func mainPage(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	writeGeneral(w, r, all)
	if admin(logged, session) {
		w.Write([]byte(`
		<form action="/users" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_name">Name: </label>
			<input type="text" id="form_name" name="name"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_name">Value: </label>
    		<input type="text" id="form_value" name="password"></p>
  		</div>
		<div class="form-row"><p>
			<input type="submit" name="but" value="Add man">
			<input type="submit" name="but" value="Change password">
			<input type="submit" name="but" value="Change name">
			<input type="submit" name="but" value="Delete man"></p>
		</div>
		</form>
		<form action="/addcard" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_card">Card: </label>
			<input type="text" id="form_card" name="card"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_info">Info: </label>
			<input type="text" id="form_info" name="info"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_name">For man: </label>
    		<input type="text" id="form_name" name="name">
			<label for="form_show">Shown: </label>
    		<input type="checkbox" id="form_show" name="shown"></p>
  		</div>
		<div class="form-row">
			<input type="submit" value="Add card">
		</div>
		</form>
		<form action="/opercard" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_card">Card number: </label>
			<input type="text" id="form_card" name="num"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_name">From man: </label>
    		<input type="text" id="form_name" name="name"></p>
  		</div>
		<div class="form-row"><p>
			<input type="submit" name="card_oper" value="Delete card">
			<input type="submit" name="card_oper" value="Make card shown">
			<input type="submit" name="card_oper" value="Make card hidden"></p>
		</div>
		</form>
		<form action="/setpics" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_name">Picture name: </label>
			<input type="text" id="form_name" name="name"></p>
		</div>
		<div class="form-row"><p>
			<label for="form_url">Url: </label>
    		<input type="text" id="form_url" name="url"></p>
  		</div>
		<div class="form-row"><p>
			<label for="form_info">Info: </label>
    		<input type="text" id="form_info" name="info"></p>
  		</div>
		<div class="form-row"><p>
			<input type="submit" name="pic_oper" value="Delete pic">
			<input type="submit" name="pic_oper" value="Create/edit pic"></p>
		</div>
		</form>
		<form action="/reload" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_saved">saved.txt </label>
			<textarea rows="3" cols="30" name="saved"></textarea></p>
		</div>
		<div class="form-row"><p>
			<input type="submit" name="load" value="Reload"></p>
		</div>
		</form>
		<form action="/download" method="post" class="reg-form">
		<div class="form-row"><p>
			<input type="submit" name="load" value="Download"></p>
		</div>
		</form>`))
	}
	w.Write([]byte(`<p>   ` + all.mainPage + `</p>`))
	if logged {
		w.Write([]byte(`
			<form action="/action" method="post" class="reg-form">
			<div class="form-row"><p>
				<input type="submit" name="action" value="Show/hide cards"></p>
			</div>
			</form>
		`))
	}
	w.Write([]byte(`
		<table bgcolor="white">
			<tr bgcolor="#29DD97">
				<th><p>Name</p></th>
				<th><p>Cards</p></th>
			</tr>`))
	all.mu.Lock()
	hiddenUrl := hiddenPic(all.cards).url
	for manname, man := range all.mans {
		if len(man.cards) > 0 {
			w.Write([]byte(`<tr><td><p>` + manname + `</p></td><td>`))
			for _, card := range man.cards {
				found, exists := all.cards[(*card).name]
				if exists {
					if card.show || (logged && manname == session.Value) ||
						admin(logged, session) {
						if card.info != "" {
							w.Write([]byte(`
								<img src="` + found.url +
								`" title="` + card.info + `">
							`))
						} else {
							w.Write([]byte(`
								<img src="` + found.url +
								`" title="` + found.info + `">
							`))
						}
					} else {
						w.Write([]byte(`
						<img src="` + hiddenUrl + `" title="No access">`))
					}
				}
			}
			w.Write([]byte(`</td></tr>`))
		}
	}
	all.mu.Unlock()
	w.Write([]byte(`</table>`))
	writeEnd(w)
}

func commPage(w http.ResponseWriter, r *http.Request, all *All) {
	_, logged := getCookies(r)
	writeGeneral(w, r, all)
	w.Write([]byte(`<form action="/send" method="post" class="reg-form">`))
	if !logged {
		w.Write([]byte(`
		<div class="form-row"><p>
			<label for="form_url">Who are you? </label>
    		<input type="text" id="form_name" name="name"></p>
  		</div>`))
	}
	w.Write([]byte(`
		<div class="form-row"><p>
			<label for="form_list">Comment: </label>
			<textarea rows="1" cols="30" name="mess"></textarea>
			<input type="submit" name="send" value="Send"></p>
		</div>
		</form>
		<p></p>`))
	all.mu.Lock()
	l := len(all.comms) - 1
	for i := range all.comms {
		w.Write([]byte(`<p><span style="color:#8B0000">[` +
			(*all.comms[l - i]).time + `]</span> `))
		w.Write([]byte(`<b>` + (*all.comms[l - i]).name + `: </b>` +
			code((*all.comms[l - i]).text) + `</p>`))
	}
	all.mu.Unlock()
	writeEnd(w)
}

func eventsPage(w http.ResponseWriter, r *http.Request, all *All) {
	writeGeneral(w, r, all)
	w.Write([]byte(all.eventsPage))
	writeEnd(w)
}

func cardsPage(w http.ResponseWriter, r *http.Request, all *All) {
	writeGeneral(w, r, all)
	w.Write([]byte(`
		<table border bgcolor="white">
			<tr bgcolor="#008B8B">
				<th><p>Card</p></th>
				<th><p>Info</p></th>
			</tr>`))
	all.mu.Lock()
	cardList := map[string]map[string]int{}
	infoList := map[string]string{}
	urlList := map[string]string{}
	for cardname, card := range all.cards {
		cardList[cardname] = map[string]int{}
		urlList[cardname] = (*card).url
		infoList[cardname] = (*card).info
	}
	for manname, man := range all.mans {
		for _, card := range man.cards {
			if (*card).show {
				_, exists := cardList[(*card).name]
				if exists {
					_, exists2 := cardList[(*card).name][manname]
					if exists2 {
						cardList[(*card).name][manname] += 1
					} else {
						cardList[(*card).name][manname] = 1
					}
				}
			}
		}
	}
	all.mu.Unlock()
	for i, nameList := range cardList {
		total := 0
		for _, amount := range nameList {
			total += amount
		}
		if total > 0 {
			w.Write([]byte(`<tr><td><img src="` + urlList[i] + `"></td>`))
			w.Write([]byte(`<td>`))
			if infoList[i] != "" {
				w.Write([]byte(`<table bgcolor="#B0E0E6">
				<tr><td><p>` + infoList[i] + `</p></td></tr></table>`))
			}
			for name, amount := range nameList {
				w.Write([]byte(`<p>` + name + `: ` + strconv.Itoa(amount) +
					` ` + howMany(amount, "card") + `</p>`))
			}
			w.Write([]byte(`<p><b>Total amount is <span style="color:#DC143C">` +
				strconv.Itoa(total) + `</span></b></p>`))
			w.Write([]byte(`</td></tr>`))
		}
	}
	w.Write([]byte(`</table>`))
	writeEnd(w)
}

//===[ADMIN_CARDS_OPERATIONS]================================================\\

func setPictures(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	name := r.FormValue("name")
	url := r.FormValue("url")
	info := r.FormValue("info")
	button := r.FormValue("pic_oper")
	if wrong(name) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	all.mu.Lock()
	if button == "Delete pic" {
		_, exists := all.cards[name]
		if exists {
			delete(all.cards, name)
		}
	} else if button == "Create/edit pic" {
		card, exists := all.cards[name]
		if exists {
			if url != "" {
				(*card).url = url
			}
			if info != "" {
				(*card).info = info
			}
		} else {
			newcard := Card {
				url:	url,
				info:	info,
			}
			all.cards[name] = &newcard
		}
	}
	all.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[USERS_OPERATIONS]======================================================\\

func users(w http.ResponseWriter, r *http.Request, all *All) {
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
	all.mu.Lock()
	_, exists := all.mans[name]
	if button == "Oke" && logged {
		_, exists = all.mans[session.Value]
		if exists {
			(*all.mans[session.Value]).password = pass
		}
	} else if admin(logged, session) {
		if button == "Delete man" && pass == "delete" && exists {
			delete(all.mans, name)
		} else if button == "Change name" && exists {
			all.mans[pass] = all.mans[name]
			delete(all.mans, name)
		} else if button == "Change password" && exists {
			all.mans[name].password = pass
		} else if button == "Add man" {
			man := Man {
				password:	pass,
				cards:		[]*MyCard{},
			}
			all.mans[name] = &man
		}
	}
	all.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

func operCard(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	name := r.FormValue("name")
	num := r.FormValue("num")
	button := r.FormValue("card_oper")
	number := -1
	if button == "No, I dont want" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if name != "" {
		numberTmp, err := strconv.Atoi(num)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		number = numberTmp - 1
	}
	all.mu.Lock()
	man, exists := all.mans[name]
	if button == "Oke" && logged {
		_, exists := all.mans[session.Value]
		if exists {
			for i, card := range all.mans[session.Value].cards {
				check := r.FormValue("shown" + strconv.Itoa(i))
				(*card).show = isOn(check)
			}
		}
	} else if exists && len(man.cards) > number && admin(logged, session) {
		if button == "Make card shown" {
			man.cards[number].show = true
		} else if button == "Make card hidden" {
			man.cards[number].show = false
		} else if button == "Delete card" {
			if len(man.cards) == 1 {
				man.cards = []*MyCard{}
			} else if number == len(man.cards) - 1 {
				man.cards = man.cards[: number]
			} else {
				man.cards = append(man.cards[: number],
					man.cards[number + 1 :]...)
			}
		}
	}
	all.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

func addCard(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	card := r.FormValue("card")
	name := r.FormValue("name")
	info := r.FormValue("info")
	show := r.FormValue("shown")
	_, existsc := all.cards[card]
	_, existsn := all.mans[name]
	if wrong(card) || !existsc || !existsn {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	mycard := MyCard {
		name:	card,
		show:	isShow(show),
		info:	info,
	}
	all.mu.Lock()
	all.mans[name].cards = append(all.mans[name].cards, &mycard)
	all.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[FULL_DATA]=============================================================\\

func download(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	all.mu.Lock()
	for name, i := range all.cards {
		w.Write([]byte(name + "(-ELEM-)" + (*i).url + "(-ELEM-)" + (*i).info))
		w.Write([]byte("(-STRING-)\n"))
	}
	w.Write([]byte("(-BLOCK-)\n\n"))
	for man, i := range all.mans {
		w.Write([]byte(man + "(-ELEM-)" + (*i).password + "(-ELEM-)"))
		for _, j := range i.cards {
			w.Write([]byte((*j).name + "(-PART-)" + isTrue((*j).show) +
				"(-PART-)" + ((*j).info)))
			w.Write([]byte("(-THING-)"))
		}
		w.Write([]byte("(-STRING-)\n"))
	}
	w.Write([]byte("(-BLOCK-)\n\n"))
	for _, i := range all.comms {
		w.Write([]byte(i.time + "(-ELEM-)" + i.name + "(-ELEM-)" + i.text))
		w.Write([]byte("(-STRING-)\n"))
	}
	w.Write([]byte("(-BLOCK-)\n\n"))
	w.Write([]byte(all.mainPage))
	w.Write([]byte("\n(-BLOCK-)\n\n"))
	w.Write([]byte(all.eventsPage))
	all.mu.Unlock()
}

func getAll(data string) (*All) {
	blocks := strings.Replace(data, "\r", "", -1)
	all := All{
		cards:	map[string]*Card{},
		mans:	map[string]*Man{},
		comms:	[]*Comm{},
	}
	parts := strings.Split(blocks, "(-BLOCK-)")
	if len(parts) < 5 {
		return &all
	}
	cardList := strings.Split(parts[0], "(-STRING-)")
	for _, i := range cardList {
		cardInfo := strings.Split(i, "(-ELEM-)")
		if len(cardInfo) < 2 {
			continue
		}
		info := ""
		if len(cardInfo) >= 3 {
			info = cardInfo[2]
		}
		card := Card {
			url:	cardInfo[1],
			info:	info,
		}
		all.cards[strings.Replace(cardInfo[0], "\n", "", -1)] = &card
	}
	nameList := strings.Split(parts[1], "(-STRING-)")
	for _, i := range nameList {
		nameInfo := strings.Split(i, "(-ELEM-)")
		if len(nameInfo) < 2 {
			continue
		}
		man := Man {
			password:	nameInfo[1],
			cards:		[]*MyCard{},
		}
		if len(nameInfo) == 3 {
			hisCards := nameInfo[2]
			eachCard := strings.Split(hisCards, "(-THING-)")
			for _, j := range eachCard {
				cardPointer := strings.Split(j, "(-PART-)")
				if len(cardPointer) < 2 {
					continue
				}
				_, exists := all.cards[cardPointer[0]]
				if !exists {
					continue
				}
				info := ""
				if len(cardPointer) >= 3 {
					info = cardPointer[2]
				}
				mycard := MyCard {
					name:	cardPointer[0],
					show:	isShow(cardPointer[1]),
					info:	info,
				}
				man.cards = append(man.cards, &mycard)
			}
		}
		all.mans[strings.Replace(nameInfo[0], "\n", "", -1)] = &man
	}
	commList := strings.Split(parts[2], "(-STRING-)")
	for _, i := range commList {
		commInfo := strings.Split(i, "(-ELEM-)")
		if len(commInfo) < 3 {
			continue
		}
		comm := Comm {
			name:	commInfo[1],
			text:	commInfo[2],
			time:	strings.Replace(commInfo[0], "\n", "", -1),
		}
		all.comms = append(all.comms, &comm)
	}
	all.mainPage = parts[3]
	all.eventsPage = parts[4]
	bot := 0
	all.bot = &bot
	return &all
}

func reload(w http.ResponseWriter, r *http.Request, all *All) {
	session, logged := getCookies(r)
	if !admin(logged, session) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	saved := r.FormValue("saved")
	tmp := getAll(saved)
	all.mu.Lock()
	all.mans = tmp.mans
	all.cards = tmp.cards
	all.comms = tmp.comms
	all.mainPage = tmp.mainPage
	all.eventsPage = tmp.eventsPage
	all.mu.Unlock()
	bot := 0
	all.bot = &bot
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[LOGIN]=================================================================\\

func allRight(name string, pass string, mans map[string]*Man) (bool) {
	if wrong(name) || wrong(pass) {
		return false
	}
	for manname, man := range mans {
		if manname == name {
			return (*man).password == pass
		}
	}
	return false
}

func loginPage(w http.ResponseWriter, r *http.Request, all *All) {
	expiration := time.Now().Add(10 * time.Hour)
	name := r.FormValue("name")
	pass := r.FormValue("password")
	all.mu.Lock()
	if allRight(name, pass, all.mans) {
		cookie := http.Cookie{
			Name:    "session_id",
			Value:   name,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	}
	all.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)
}

//===[USER_ACTIONS]==========================================================\\

func actionPage(w http.ResponseWriter, r *http.Request, all *All) {
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
		writeGeneral(w, r, all)
		w.Write([]byte(`
		<form action="/opercard" method="post" class="reg-form">
		<table border bgcolor="white">
			<tr bgcolor="#B22222">
				<th><p>Card</p></th>
				<th><p>Shown</p></th>
			</tr>`))
		all.mu.Lock()
		for i, card := range all.mans[session.Value].cards {
			thiscard, exists := all.cards[(*card).name]
			if !exists {
				continue
			}
			w.Write([]byte(`<tr>
				<td><img src=` + (*thiscard).url + `></td>
				<td><div class="form-row">`))
			is := strconv.Itoa(i)
			if (*card).show {
				w.Write([]byte(`
					<input type="checkbox" id="check` + is +
					`" name="shown` +
					is + `" checked="checked">
				`))
			} else {
				w.Write([]byte(`
					<input type="checkbox" id="check` + is +
					`" name="shown` + is + `">
				`))
			} 
			w.Write([]byte(`</div></td></tr>`))
		}
		all.mu.Unlock()
		w.Write([]byte(`
			</table>
			<div class="form-row"><p>
				<input type="submit" name="card_oper" value="Oke">
				<input type="submit" name="card_oper" value="No, I dont want">
			</p></div>
			</form>`))
		writeEnd(w)
		return
	}
	// else "Change Password"
	writeGeneral(w, r, all)
	w.Write([]byte(`
		<form action="/users" method="post" class="reg-form">
		<div class="form-row"><p>
			<label for="form_passnew">New password: </label>
    		<input type="password" id="form_passnew" name="password"></p>
  		</div>
		<div class="form-row"><p>
			<label for="form_passnew2">New password again: </label>
    		<input type="password" id="form_passnew2" name="password2"></p>
  		</div>
		<div class="form-row"><p>
			<input type="submit" name="but" value="Oke">
			<input type="submit" name="but" value="No, I dont want"></p>
		</div>
		</form>
	`))
	writeEnd(w)
}

func send(w http.ResponseWriter, r *http.Request, all *All) {
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
	comm := Comm {
		name:	name,
		text:	mess,
		time:	strings.Split(time.Now().String(), ".")[0],
	}
	all.mu.Lock()
	all.comms = append(all.comms, &comm)
	all.mu.Unlock()
	http.Redirect(w, r, "/comments", http.StatusFound)
}

//===[BOT]===================================================================\\

func getBear(w http.ResponseWriter, r *http.Request, all *All) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!doctype html><html><body><p>TEST!</p></body></html>`))
	all.mu.Lock()
	if *all.bot > 0 {
		*all.bot = 0
	} else {
		*all.bot -= 1
	}
	if *all.bot < -2 {
		*all.bot = 0
		all.mu.Unlock()
		log.Println("HELP!")
		go sendCat(w, r, all)
		all.mu.Lock()
	}
	all.mu.Unlock()
	log.Println("frombot-Got ", *all.bot)
}

func sendCat(w http.ResponseWriter, r *http.Request, all *All) {
	for {
		time.Sleep(3 * time.Minute)
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
				all.mu.Lock()
				*all.bot = 1
				all.mu.Unlock()
				log.Println("tobot-Done ", *all.bot)
			}
		} else {
			log.Println("request error" + err.Error())
		}
	}
}

//===[MAIN]==================================================================\\

func main() {
	data, _ := ioutil.ReadFile("saved.txt")
	all := getAll(string(data))

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginPage(w, r, all)
	})

	http.HandleFunc("/action", func(w http.ResponseWriter, r *http.Request) { 
		actionPage(w, r, all)
	})

	http.HandleFunc("/addcard", func(w http.ResponseWriter, r *http.Request) {
		addCard(w, r, all)
	})

	http.HandleFunc("/opercard", func(w http.ResponseWriter, r *http.Request) {
		operCard(w, r, all)
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		users(w, r ,all)
	})

	http.HandleFunc("/setpics", func(w http.ResponseWriter, r *http.Request) {
		setPictures(w, r, all)
	})

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		reload(w, r, all)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		download(w, r, all)
	})

	http.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		commPage(w, r, all)
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		send(w, r, all)
	})

	http.HandleFunc("/contests", func(w http.ResponseWriter, r *http.Request) {
		eventsPage(w, r, all)
	})

	http.HandleFunc("/cards", func(w http.ResponseWriter, r *http.Request) {
		cardsPage(w, r, all)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mainPage(w, r, all)
	})
	
	http.HandleFunc("/sendbot", func(w http.ResponseWriter, r *http.Request) {
		sendCat(w, r, all)
	})

	http.HandleFunc("/getbot", func(w http.ResponseWriter, r *http.Request) {
		getBear(w, r, all)
	})

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	http.ListenAndServe(":"+port, nil)

	log.Println("starting server at :8080")
	//http.ListenAndServe(":8080", nil)
}
