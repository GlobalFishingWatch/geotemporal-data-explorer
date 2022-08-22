# Geotemporal-data-explorer

## Install

```bash

curl -o- https://storage.googleapis.com/geotemporal-data-explorer-releases/install-latest.sh | sudo bash

```

## Run Serve

```

geotemporal-data-explorer serve

````

Available options:

* -p: Port to listen (default 8080)
* --gee-account-file=< path > : Path to the Service account to use Google Earth Engine
* --gfw-token=< token >: Token of Global Fishing Watch to use data of it. You can obtain your token [here](https://globalfishingwatch.org/ocean-engine/tokens)
