package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// currencies and their exchange rates
var exchangeRates = map[string]float64{
	"btc": 0.0, "eth": 0.0, "ltc": 0.0, "bch": 0.0, "bnb": 0.0,
	"eos": 0.0, "xrp": 0.0, "xlm": 0.0, "usd": 0.0, "aed": 0.0,
	"ars": 0.0, "aud": 0.0, "bdt": 0.0, "bhd": 0.0, "bmd": 0.0,
	"brl": 0.0, "cad": 0.0, "chf": 0.0, "clp": 0.0, "cny": 0.0,
	"czk": 0.0, "dkk": 0.0, "eur": 0.0, "gbp": 0.0, "hkd": 0.0,
	"huf": 0.0, "idr": 0.0, "ils": 0.0, "inr": 0.0, "jpy": 0.0,
	"krw": 0.0, "kwd": 0.0, "lkr": 0.0, "mmk": 0.0, "mxn": 0.0,
	"myr": 0.0, "nok": 0.0, "nzd": 0.0, "php": 0.0, "pkr": 0.0,
	"pln": 0.0, "rub": 0.0, "sar": 0.0, "sek": 0.0, "sgd": 0.0,
	"thb": 0.0, "try": 0.0, "twd": 0.0, "vef": 0.0, "zar": 0.0,
	"xdr": 0.0, "xag": 0.0, "xau": 0.0, "trtl": 1.0, // 1 TRTL = 1 TRTL
}

const id = "turtlecoin"

// payload for coingecko
const currencies = "btc,eth,ltc,bch,bnb,eos,xrp,xlm,usd,aed,ars,aud,bdt,bhd," +
	"bmd,brl,cad,chf,clp,cny,czk,dkk,eur,gbp,hkd,huf,idr,ils,inr,jpy,krw,kwd," +
	"lkr,mmk,mxn,myr,nok,nzd,php,pkr,pln,rub,sar,sek,sgd,thb,try,twd,vef,zar," +
			"xdr,xag,xau"
const maxAge = 300 // time in seconds to update exchange rates

func init() {
	ch := make(chan bool, 1)
	go getExchangeRates(ch)
	<-ch
}

//getExchangeRates - gets the exchange rate from coingecko
func getExchangeRates(ch chan bool) error {
	for ; ; time.Sleep(time.Second * maxAge) {
		var jsresp map[string]interface{}
		resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=" + id + "&vs_currencies=" + currencies)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err = json.Unmarshal(data, &jsresp); err != nil {
			panic(err)
		}

		rates := jsresp[id].(map[string]interface{})
		for cur, val := range rates {
			exchangeRates[cur] = val.(float64)
		}
		ch <- true
	}
}

/*
func main() {
	ch := make(chan bool, 1)
	go getExchangeRates(ch)
	fmt.Println(<-ch)
	time.Sleep(time.Second * 2)
	item_cost := 15.0
	trtl_cost := item_cost / exchangeRates["usd"]
	fmt.Println("item cost usd:", item_cost)
	fmt.Println("item cost trtl:", trtl_cost)
}
*/
