package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	//"strings"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/oschwald/geoip2-golang"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func iphandler(w http.ResponseWriter, r *http.Request) {

	ip, port, _ := net.SplitHostPort(r.RemoteAddr)
	//session-name adindaki sessioni tarayicidan aliyoruz
	session, _ := store.Get(r, "session-name")

	//Eger sessionda client adinda bir veri yoksa bu if e giriyor
	if session.Values["Client"] == nil {
		//Dogru random sayi uretmesi icin seedliyoruz bunu yapmazsak hep ayni sayiyi uretiyor
		rand.Seed(time.Now().UnixNano())

		// client-xxxxx 0 dan 10000 e kadar sayisi olan bir kullanici id si olusturuyor
		var clientid = "client-" + strconv.Itoa(rand.Intn(10000))
		//Sessionda Client adli yere bu id yi tanimliyoruz
		session.Values["Client"] = clientid

		//Sessionda yaptigimiz degisikliklerden sonra sessionu kaydediyoruz
		err := session.Save(r, w)

		//Eger session kayit edilirken hata olursa bir http error olusturuyoruz
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//ip.gohtml dosyasini okuyoruz
	t, err := template.ParseFiles("index.html")
	//Eger bu dosya yoksa hata oluusturuyor
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type Data struct {
		Ip      string
		Client  string
		Country string
		City    string
		Port    string
	}
	dataToSend := Data{
		Ip:      ip,
		Client:  session.Values["Client"].(string),
		Country: GetCountry(ip),
		City:    GetCity(ip),
		Port:    port,
	}

	err = t.Execute(w, dataToSend)
	if err != nil {
		panic(err)
	}
}

func addressHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	type Data struct {
		Ip      string
		Country string
		City    string
	}
	dataToSend := Data{
		Ip:      address,
		Country: GetCountry(address),
		City:    GetCity(address),
	}

	t, err := template.ParseFiles("index.html")
	//Eger bu dosya yoksa hata oluusturuyor
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = t.Execute(w, dataToSend)
	if err != nil {
		panic(err)
	}
}

func addressJsonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	type Data struct {
		Ip      string
		Country string
		City    string
	}
	dataToSend := Data{
		Ip:      address,
		Country: GetCountry(address),
		City:    GetCity(address),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataToSend)
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("x-original-forwarded-for")
	fmt.Printf("%v", forwarded)
	if forwarded != "" {

		return forwarded
	}
	return r.RemoteAddr
}
func GetCountry(ip string) string {

	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ipcheck := net.ParseIP(ip)
	fmt.Println(ip)
	fmt.Println(ipcheck)
	record, err := db.City(ipcheck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(record.Country)
	return record.Country.Names["en"]
}

func GetCity(ip string) string {

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ipcheck := net.ParseIP(ip)
	fmt.Println(ip)
	fmt.Println(ipcheck)
	record, err := db.City(ipcheck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(record.City)
	return record.City.Names["en"]
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", iphandler)
	r.HandleFunc("/ip/{address}", addressHandler)
	r.HandleFunc("/ip/{address}/json", addressJsonHandler)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
