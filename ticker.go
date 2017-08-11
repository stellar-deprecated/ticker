package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"log"
	"strings"
	"github.com/BurntSushi/toml"
	"os"
	"strconv"
)

type currencies struct {
    Title            string
    Pair             []pair
}    

type pair struct {
    Name  			 string
    Base             string
    Base_issuer      string
    Counter          []string
    Counter_issuer   []string
}

type OffersPage struct {
	Links struct {
		Self Link `json:"self"`
		Next Link `json:"next"`
		Prev Link `json:"prev"`
	} `json:"_links"`
	Embedded struct {
		Records []Offer `json:"records"`
	} `json:"_embedded"`
}

type Link struct {
	Href   string `json:"href"`
}

type Offer struct {

	Sold    string `json:"sold_amount"`
	Bought  string `json:"bought_amount"`
	When    time.Time `json:"created_at"`
}

type ToFile struct {
	Name			string
	Price			float64
	Base_Volume		float64
	Counter_Volume	float64
}

func main() {
	var config currencies
    var to_write []ToFile
    var aggregate_data []Offer
    var blank []Offer
// Read the configuration file which determines currency pairs
    if _, err := toml.DecodeFile("config.toml", &config); err != nil {
        panic(err)
    }
   
    for _, pair:= range config.Pair { //iterate through currency pairs
    		aggregate_data = blank //reset the data struct
    		fmt.Println("\n------------------------------------------", pair.Name)
        for index, _:= range pair.Counter_issuer { //iterate through issuers of given currency
            link := get_link(pair, index, 200) 
            aggregate_data = append(aggregate_data, get_book(link)...) //struct contains all transactions for given currency pair
      	}
      	
      	price := get_price(pair, 0)
        buyer_volume, seller_volume := get_volume(aggregate_data)

        fmt.Println(price, buyer_volume, seller_volume)
        to_write = append(to_write, ToFile{Name: pair.Name, Price: price, Base_Volume: buyer_volume, Counter_Volume: seller_volume})
    }
// Write to file
    path := "exchange.json"
    fo, err := os.Create(path)
    if err != nil {
        panic(err)
    }
    defer fo.Close()
    e := json.NewEncoder(fo)
    if err:= e.Encode(to_write); err != nil {
        panic(err)
    }    
}

func get_link(currency pair, index int, limit int) string {
    var link string
    switch { 
    case currency.Counter[index] == "XLM":
        link = fmt.Sprintf("https://horizon.stellar.org/order_book/trades?selling_asset_type=credit_alphanum4&selling_asset_code=%s&selling_asset_issuer=%s&buying_asset_type=native&limit=%d&order=desc", currency.Base, currency.Base_issuer, limit)
    case currency.Base == "XLM":
        link = fmt.Sprintf("https://horizon.stellar.org/order_book/trades?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_code=%s&buying_asset_issuer=%s&limit=%d&order=desc", currency.Counter[index], currency.Counter_issuer[index], limit)
    default:
        link = fmt.Sprintf("https://horizon.stellar.org/order_book/trades?selling_asset_type=credit_alphanum4&selling_asset_code=%s&selling_asset_issuer=%s&buying_asset_type=credit_alphanum4&buying_asset_code=%s&buying_asset_issuer=%s&limit=%d&order=desc",currency.Base, currency.Base_issuer, currency.Counter[index], currency.Counter_issuer, limit)
    }
    return link
}

func get_book(link string) []Offer{
	var sub_resp OffersPage
	var data OffersPage
	var result int
	var upper int
	
	for {
		fmt.Println(link)
		sub_resp = get_request(link) // perform get request on link
	    upper = len(sub_resp.Embedded.Records) //the total number of trades on the page
	    result = this_page(sub_resp, upper-1) //binary search to find oldest trade to happen within 24 hours

	    switch{
		case result == -2: //next page
			data.Embedded.Records = append(data.Embedded.Records, sub_resp.Embedded.Records...)
	        link = sub_resp.Links.Prev.Href
	        link = strings.Replace(link, "\u0026", "&", -1)
	    case result == -1: //no relevant trades on this page
	        return data.Embedded.Records
	    default: //copy all trades that occured within past 24 hours
	    	data.Embedded.Records = append(data.Embedded.Records, sub_resp.Embedded.Records[0:(result+1)]...)
	    	return data.Embedded.Records    
	    }
	}
}

func get_request(link string) OffersPage { //performs get request
	var sub_resp OffersPage

	tr := &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    30 * time.Second,
	DisableCompression: true,
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(link)
	if err != nil {
		log.Fatalln("unable to make request: ", err)
    }
	decodeResponse(resp, &sub_resp)

	return sub_resp
}

func decodeResponse(resp *http.Response, object interface{}) (err error) { //puts response from get request in given struct
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		panic(resp.StatusCode)
	}
	err = decoder.Decode(&object)
	if err != nil {
		return
	}
	return
}

func this_page(sub_resp OffersPage, upper int) int{ //performs a binary search through a page to determine how many trades occured in last 24 hours
	var index int
	lower := 0
	utc := time.Now().UTC()
    yesterday := utc.AddDate(0, 0, -1)
    switch {
 	case yesterday.Before(sub_resp.Embedded.Records[upper-1].When): //all the trades on this page, move to next page
    	return -2
 	case yesterday.After(sub_resp.Embedded.Records[lower].When): // none of the trades on this page
 		return -1
 	default:
        for {
            index = (upper + lower + 1)/2
            if yesterday.After(sub_resp.Embedded.Records[index].When){
                if (upper-lower) <= 1{
                	return lower
                }
           		upper = index
            } else{
                if (upper-lower) <= 1{
                    return upper
                }
            	lower = index
            }
        }
	}
	return -3
}

func get_price(Pair pair, index int) float64{
	var sub_resp OffersPage
	var err error

	link := get_link(Pair, index, 1)

	sub_resp = get_request(link)

	bought, err := strconv.ParseFloat(sub_resp.Embedded.Records[0].Bought, 64)
	sold, err := strconv.ParseFloat(sub_resp.Embedded.Records[0].Sold, 64)
	if err!=nil {
		panic(err)
	}
	return bought/sold
}

func get_volume(data []Offer) (float64, float64) {
    var err error
    var bought, sold float64
    seller_volume := 0.0
    buyer_volume := 0.0
    
    for _, item := range data {
        bought, err = strconv.ParseFloat(item.Sold, 64)
        sold, err = strconv.ParseFloat(item.Bought, 64)
        if err != nil {
        	panic(err)
        }
        buyer_volume += bought 
        seller_volume += sold 
    }
    return buyer_volume, seller_volume 
}
