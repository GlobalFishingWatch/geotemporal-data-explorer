# Geotemporal-data-explorer

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


### Google Earth Engine Account

To obtain your account, first you need to register yourself [here](https://earthengine.google.com/) 
When you already have your account, you need to obtain a Service account with your GEE account. You can obtain it
following the steps of this [link](https://developers.google.com/earth-engine/guides/service_account)

