# Geotemporal-data-explorer

Geotemporal-data-explorer is a tool to research the Google Earth Engine datasets, Global Fishing Watch datasets and your own datasets using the 4wings technology. You can know more about 4wings in this [presentation](https://docs.google.com/presentation/d/1OJCg2zJp0zEVcYJ6Z4ePywy0oO59FnZ2aPUXp7LB4sc/edit?usp=sharing)

## Install

Only available for linux with intel chips and OSX with ARM chips. Working in support more platforms.

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


## Add new Google Earth Engine dataset

When you run the application for the first time, the application creates a folder called data in the path where you ran it. Inside there is a `datasets.json` file. Inside this file you can add the datasets you want.
A GEE dataset is composed of the following fields:

* id: Id of the dataset. Should be unique
* name: Name of the dataset.
* description: Description of the dataset
* startDate: Start date of the dataset in ISO format.
* endDate: End date of the dataset in ISO format.
* configuration:
  * band: Name of the band that contains the data.
  * intervals: Array with the intervals supported for the dataset. It's the granularity of the dataset. Supported values: hour, day, month, year
  * images: Array with 4 positions. Each position should have the name of the image that contains the data in each granularity. 
    * Position 0: ImageCollection for hourly data. If it does not have, the value should be "".
    * Position 1: ImageCollection for dayly data. If it does not have, the value should be "".
    * Position 2: ImageCollection for monthly data. If it does not have, the value should be "".
    * Position 3: ImageCollection for yearly data. If it does not have, the value should be "".
  * scale: Scale of the data in the dataset. If it does not have, the value should be 1.
  * offset: Offset if the data. If it does not have, the value should be 0.
  * min: Minimum value of the data
  * max: Maximum value of the data
  
Example:

We are going to add the dataset [MODIS_006_MYD10A](https://developers.google.com/earth-engine/datasets/catalog/MODIS_006_MYD10A1). This is a daily dataset with data from 2002-07-04 to 2022-09-29.

The json of the dataset is:

```json

 {
    "id": "public-global-aqua-snow-cover",
    "name": "Aqua Snow Cover Daily Global 500m",
    "description": "Aqua Snow Cover Daily Global 500m",
    "source": "GEE",
    "type": "4wings",
    "startDate": "2002-07-04T00:00:00Z",
    "endDate": "2022-09-29T23:00:00Z",
    "configuration": {
      "intervals": ["day"],
      "scale": 1,
      "offset": 0,
      "min": 0,
      "max": 100,
      "band": "NDSI_Snow_Cover",
      "images": ["", "MODIS/006/MYD10A1", "", ""]
    }
  }

```

In this case, as the dataset has daily data, we only add day interval and informed the image for daily data (Position 1 in the array). The scale, offset, min and max values are documented in the [GEE webpage of the dataset](https://developers.google.com/earth-engine/datasets/catalog/MODIS_006_MYD10A1#bands)

We add the json object inside of the brackets (`[]`) of the `datasets.json`. If the file, already contains other datasets, we will add the new object, after of the current datasets.

In a empty `datasets.json` file, the file will look like:

```json
[
 {
    "id": "public-global-aqua-snow-cover",
    "name": "Aqua Snow Cover Daily Global 500m",
    "description": "Aqua Snow Cover Daily Global 500m",
    "source": "GEE",
    "type": "4wings",
    "startDate": "2002-07-04T00:00:00Z",
    "endDate": "2022-09-29T23:00:00Z",
    "configuration": {
      "intervals": ["day"],
      "scale": 1,
      "offset": 0,
      "min": 0,
      "max": 100,
      "band": "NDSI_Snow_Cover",
      "images": ["", "MODIS/006/MYD10A1", "", ""]
    }
  }
]
```
