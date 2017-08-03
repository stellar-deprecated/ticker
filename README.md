# ticker #

A script to generate a json file for stellar stats

## Input ##

config.toml - ticker.go pulls currency pairs from here and performs calculations

### Formatting ###

The top of the config.toml file should have a Title line that resembles:
    Title = "Currency Pairs"

Below the title, currency pairs can be listed like so:
    [[pair]]
    name = "XLM_BTC"
    base = "XLM"
    base_issuer = "native"
    counter = ["BTC"]
    counter_issuer = ["GATEMHCCKCY67ZUCKTROYN24ZYT5GK4EQZ65JJLDHKHRUZI3EUEKMTCH"]

If you wish to aggregate the volume for mutliple tokens and issuers, you can enter the values as a comma seperated list. Example below:
    [[pair]]
    name = "XLM_BTC"
    base = "XLM"
    base_issuer = "native"
    counter = ["BTC", "EURT"]
    counter_issuer = ["GATEMHCCKCY67ZUCKTROYN24ZYT5GK4EQZ65JJLDHKHRUZI3EUEKMTCH", "GAP5LETOV6YIE62YAM56STDANPRDO7ZFDBGSNHJQIYGGKSMOZAHOOS2S"]

The asset "BTC" corresponds to the issuer starting with "GAT", and the asset "EURT" corresponds to the issuer starting with "GAP". To add more values, continue the comma seperated list. Every counter must have its corresponding issuer listed at the appropriate place in counter_issuer. Ie. The third item in counter must correspond to the third item in counter_issuer. 

NOTE: You can only have one asset for base, but may have as many assets as you would like for counter.

## To run: ##

Requires toml
* go get github.com/BurntSushi/toml
* go install github.com/BurntSushi/toml

go run ticker.go

## output ## 

exchange.json
