# Geotemporal-data-explorer

Geotemporal-data-explorer is a tool to research the Google Earth Engine datasets, Global Fishing Watch datasets and your own datasets using the 4wings technology. You can know more about 4wings in this [presentation](https://docs.google.com/presentation/d/1OJCg2zJp0zEVcYJ6Z4ePywy0oO59FnZ2aPUXp7LB4sc/edit?usp=sharing)

## Install

```bash
sudo curl -o- https://storage.googleapis.com/geotemporal-data-explorer-releases/install-last.sh | sudo bash
```

## Run Serve

```

geotemporal-data-explorer serve

````

Available options:

* -p: Port to listen (default 8080)
* --gee-account-file=< path > : Path to the Service account to use Google Earth Engine. To obtain your account, follow the steps of [here](#google-earth-engine-account)
* --gfw-token=< token >: Token of Global Fishing Watch to use data of it. You can obtain your own token [here](https://globalfishingwatch.org/ocean-engine/tokens)

Remember, in OSX, the first time that you run it, you will need to give permissions to the app in `Security & Privacy`

### Google Earth Engine Account

To obtain your account, first you need to register yourself [here](https://earthengine.google.com/) 
When you already have your account, you need to obtain a Service account with your GEE account. You can obtain it
following the steps of this [link](https://developers.google.com/earth-engine/guides/service_account)

